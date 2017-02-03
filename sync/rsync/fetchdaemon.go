package rsync

import (
	"fmt"
	"bytes"
	"os"
	"time"
	"strings"
	"io/ioutil"

	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/osapi"
	kexec "github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
)

func init() {
	RfetchDaemon = NewRsyncDaemonFetch()
}

func NewRsyncDaemonFetch() *RsyncDaemonFetch {
	d := &RsyncDaemonFetch{}
	d.executor = kexec.NewLocal()
	return d
}

const startTimeout = 10

const startDaemon = `
TMP=${TMP:-/tmp}
CONFIG=$(echo -n "" > ${TMP}/%[1]s && echo ${TMP}/%[1]s)
PID=$(echo -n "" > ${TMP}/%[2]s && echo ${TMP}/%[2]s)
rm $PID
printf "pid file = ${PID}\n[root]\n  path = /\n  use chroot = no\n  read only = no" > $CONFIG
rsync --no-detach --daemon --config=${CONFIG} --address=127.0.0.1 --port=%[3]d
`

const killDaemon = `set -e
TMP=${TMP:-/tmp}
PID=${TMP}/%[1]s
kill $(cat ${PID})
`
const checkRunningDeamon = `set -e
TMP=${TMP:-/tmp}
PID=${TMP}/%[1]s
ls ${PID}
`

type RsyncDaemonFetch struct {
	executor kexec.Executor
	kscmd    kexec.KSCommand
}

func (r RsyncDaemonFetch) Fetch(kubeConfigKey string, environment string, pod string, filePath string) error {
	configFile := "rsyncd.conf"
	pidFile := "rsyncd-pid.conf"
	port := 9999

	r.kscmd = kexec.KSCommand{}
	r.kscmd.KubeConfigKey = kubeConfigKey
	r.kscmd.Environment = environment
	r.kscmd.Pod = pod
	r.kscmd.Stderr = ioutil.Discard
	r.kscmd.Stdout = ioutil.Discard

	errChan := r.StartDaemon(configFile, pidFile, port)

	err := r.WaitForDaemon(pidFile, errChan)
	if err != nil {
		return err
	}
	defer r.KillDaemon(pidFile)

	stopChan := r.StartPortForward("9999:9999")
	defer r.StopPortForward(stopChan)

	args := []string{
		"-zrlptDv",
		"--blocking-io",
		"--force",
		`--exclude=".*"`,
		fmt.Sprintf(`--exclude-from=%s`, SyncExcluded),
		"--",
	}

	if filePath == "" {
		cplogs.V(5).Infoln("fetching all files")
		args = append(args, r.GetRsyncURL(port, "root", "app"))
	} else {
		cplogs.V(5).Infof("fetching specified file %s", filePath)
		args = append(args, r.GetRsyncURL(port, "root", "app/"+filePath))
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}
	args = append(args, currentDir)

	cplogs.V(5).Infof("rsync arguments: %s", args)
	cplogs.Flush()

	scmd := osapi.SCommand{}
	scmd.Name = "rsync"
	scmd.Stdin = os.Stdin
	scmd.Stdout = os.Stdout
	scmd.Stderr = os.Stderr

	return osapi.CommandExecL(scmd, args...)
}

func (r RsyncDaemonFetch) StartDaemon(configFile string, pidFile string, port int) *chan error {
	r.kscmd.Stdin = bytes.NewBufferString(fmt.Sprintf(startDaemon, configFile, pidFile, port))
	errChan := make(chan error, 1)
	go func() {
		err := r.executor.StartProcess(r.kscmd, "sh")
		if err != nil {
			errChan <- err
		}
	}()
	return &errChan
}

func (r RsyncDaemonFetch) KillDaemon(pidFile string) error {
	r.kscmd.Stdin = bytes.NewBufferString(fmt.Sprintf(killDaemon, pidFile))
	err := r.executor.StartProcess(r.kscmd, "sh")
	if err != nil {
		cplogs.V(4).Infof("error when killing rsync daemon with pid file: %s, error %s", pidFile, err.Error())
	}
	return err
}

func (r RsyncDaemonFetch) WaitForDaemon(pidFile string, errChan *chan error) error {
	r.kscmd.Stdin = bytes.NewBufferString(fmt.Sprintf(checkRunningDeamon, pidFile))
	startTime := time.Now()
	for {
		if time.Since(startTime) > (startTimeout * time.Second) {
			cplogs.V(4).Infof("rsync deamon start timeout")
			return fmt.Errorf("Rsync deamon start timeout afer waiting %d seconds", startTimeout)
		}
		err := r.executor.StartProcess(r.kscmd, "sh")
		if err == nil {
			break
		}
		if len(*errChan) > 0 {
			return <-*errChan
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (r RsyncDaemonFetch) GetRsyncURL(port int, label string, path string) string {
	return fmt.Sprintf("rsync://127.0.0.1:%d/%s/%s", port, label, strings.TrimPrefix(path, "/"))
}

func (r RsyncDaemonFetch) StartPortForward(ports string) (stopChan chan bool) {
	killProcess := make(chan bool, 1)
	cplogs.V(5).Infoln("starting port forward in the background")
	go func() {
		err := kubectlapi.Forward(r.kscmd.KubeConfigKey, r.kscmd.Environment, r.kscmd.Pod, ports, killProcess)
		if err != nil {
			cplogs.V(4).Infoln("an error occured during port forward error: %s", err.Error())
		}
	}()
	return stopChan
}

func (r RsyncDaemonFetch) StopPortForward(stopChan chan bool) {
	cplogs.V(5).Infoln("stopping port forward")
	close(stopChan)
}
