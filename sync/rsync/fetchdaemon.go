package rsync

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	kexec "github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"strconv"
	"path/filepath"
	"runtime"
)

func init() {
	RfetchDaemon = NewRsyncDaemonFetch()
}

func NewRsyncDaemonFetch() *RsyncDaemonFetch {
	d := &RsyncDaemonFetch{}
	d.remoteRsync = NewRemoteRsyncDeamon()
	return d
}

const (
	commandTimeout     = 10
	rsyncConfigSection = "root"
	portForward        = 31986
	configFile         = "rsyncd.conf"
	pidFile            = "rsyncd-pid.conf"
)
const startDaemon = `
TMP=${TMP:-/tmp}
CONFIG=$(echo -n "" > ${TMP}/%[1]s && echo ${TMP}/%[1]s)
PID=$(echo -n "" > ${TMP}/%[2]s && echo ${TMP}/%[2]s)
rm $PID
printf "pid file = ${PID}\n[root]\n  path = /\n  use chroot = no\n  read only = no\n  uid = root\n  gid = root\n"  > $CONFIG
rsync --no-detach --daemon --config=${CONFIG} --address=127.0.0.1 --port=%[3]d
`

const killDaemon = `set -e
TMP=${TMP:-/tmp}
PID=${TMP}/%[1]s
kill $(cat ${PID})
`
const checkRsyncDaemon = `set -e
TMP=${TMP:-/tmp}
PID=${TMP}/%[1]s
ls ${PID}
`

type RsyncDaemonFetch struct {
	remoteRsync *RemoteRsyncDeamon
}

func (r RsyncDaemonFetch) Fetch(kubeConfigKey string, environment string, pod string, filePath string) error {
	kscmd := kexec.KSCommand{}
	kscmd.KubeConfigKey = kubeConfigKey
	kscmd.Environment = environment
	kscmd.Pod = pod
	kscmd.Stderr = ioutil.Discard
	kscmd.Stdout = ioutil.Discard
	r.remoteRsync.SetKSCommand(kscmd)

	errChan := r.remoteRsync.StartDaemon(configFile, pidFile, portForward)

	err := r.remoteRsync.WaitForDaemon(pidFile, errChan)
	if err != nil {
		return err
	}
	defer r.remoteRsync.KillDaemon(pidFile)

	stopChan, err := r.remoteRsync.StartPortForward(strconv.Itoa(portForward))
	if err != nil {
		return err
	}
	defer r.remoteRsync.StopPortForward(stopChan)
	if err != nil {
		return err
	}
	args := []string{
		"-zrltDv",
		"--blocking-io",
		"--force",
		`--exclude=.*`,
		fmt.Sprintf(`--exclude-from=%s`, SyncExcluded),
		"--",
	}

	if filePath == "" {
		cplogs.V(5).Infoln("fetching all files")
		args = append(args, r.remoteRsync.GetRsyncURL(portForward, rsyncConfigSection, "app"))
	} else {
		cplogs.V(5).Infof("fetching specified file %s", filePath)
		args = append(args, r.remoteRsync.GetRsyncURL(portForward, rsyncConfigSection, "app/"+filePath))
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		currentDir = convertWindowsPath(currentDir)
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

type RemoteRsyncDeamon struct {
	executor kexec.Executor
	kscmd    kexec.KSCommand
}

func NewRemoteRsyncDeamon() *RemoteRsyncDeamon {
	r := &RemoteRsyncDeamon{}
	r.executor = kexec.NewLocal()
	return r
}

func (r *RemoteRsyncDeamon) SetKSCommand(kscmd kexec.KSCommand) {
	r.kscmd = kscmd
}

func (r RemoteRsyncDeamon) StartDaemon(configFile string, pidFile string, port int) *chan error {
	start := fmt.Sprintf(startDaemon, configFile, pidFile, port)
	r.kscmd.Stdin = bytes.NewBufferString(start)
	errChan := make(chan error, 1)
	go func() {
		cplogs.V(5).Infof("Executing remotely script:\n%s\n", start)
		err := r.executor.StartProcess(r.kscmd, "sh")
		if err != nil {
			errChan <- err
		}
	}()
	return &errChan
}

func (r RemoteRsyncDeamon) KillDaemon(pidFile string) error {
	stop := fmt.Sprintf(killDaemon, pidFile)
	cplogs.V(5).Infof("Executing remotely script:\n%s\n", stop)
	r.kscmd.Stdin = bytes.NewBufferString(stop)
	err := r.executor.StartProcess(r.kscmd, "sh")
	if err != nil {
		cplogs.V(4).Infof("error when killing rsync daemon with pid file: %s, error %s", pidFile, err.Error())
	}
	return err
}

func (r RemoteRsyncDeamon) WaitForDaemon(pidFile string, errChan *chan error) error {
	check := fmt.Sprintf(checkRsyncDaemon, pidFile)
	r.kscmd.Stdin = bytes.NewBufferString(check)
	startTime := time.Now()
	for {
		if time.Since(startTime) > commandTimeout*time.Second {
			cplogs.V(4).Infof("rsync deamon start timeout")
			return fmt.Errorf("Rsync deamon start timeout afer waiting %d seconds", commandTimeout)
		}
		cplogs.V(5).Infof("Executing remotely script:\n%s\n", check)
		err := r.executor.StartProcess(r.kscmd, "sh")
		if err == nil {
			break
		}
		if len(*errChan) > 0 {
			return <-*errChan
		}
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func (r RemoteRsyncDeamon) GetRsyncURL(port int, label string, path string) string {
	return fmt.Sprintf("rsync://127.0.0.1:%d/%s/%s/", port, label, strings.TrimPrefix(path, "/"))
}

func (r RemoteRsyncDeamon) StartPortForward(port string) (*chan bool, error) {
	killProcess := make(chan bool, 1)
	cplogs.V(5).Infoln("starting port forward in the background")
	go func() {
		err := kubectlapi.Forward(r.kscmd.KubeConfigKey, r.kscmd.Environment, r.kscmd.Pod, port+":"+port, killProcess)
		if err != nil {
			cplogs.V(4).Infoln("an error occured during port forward error: %s", err.Error())
		}
	}()

	//wait until the port is open
	startTime := time.Now()
	for {
		if time.Since(startTime) > commandTimeout*time.Second {
			return nil, fmt.Errorf("port forwarding timeout after %d seconds", commandTimeout)
		}
		cplogs.V(5).Infof("verifying if %s is open", "127.0.0.1:"+port)
		conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, 2*time.Second)
		if err == nil {
			cplogs.V(5).Infof("%s is open", "127.0.0.1:"+port)
			conn.Close()
			break
		}
		cplogs.V(5).Infof("%s is still closed, retrying..", "127.0.0.1:"+port)
		time.Sleep(100 * time.Millisecond)
	}
	return &killProcess, nil
}

func (r RemoteRsyncDeamon) StopPortForward(stopChan *chan bool) {
	if stopChan == nil {
		cplogs.V(5).Infoln("was not possible to stop the port forwarding")
		return
	}
	cplogs.V(5).Infoln("stopping port forwarding")
	close(*stopChan)
}

// convertWindowsPath converts a windows native path to a path that can be used by
// the rsync command in windows.
// It can take one of three forms:
// 1 - relative to current dir or relative to current drive
//     \mydir\subdir or subdir
//     For these, it's only sufficient to change '\' to '/'
// 2 - absolute path with drive
//     d:\mydir\subdir
//     These need to be converted to /cygdrive/<drive-letter>/rest/of/path
// 3 - UNC path
//     \\server\c$\mydir\subdir
//     For these it should be sufficient to change '\' to '/'
func convertWindowsPath(path string) string {
	// If the path starts with a single letter followed by a ":", it needs to
	// be converted /cygwin/<drive>/path form
	parts := strings.SplitN(path, ":", 2)
	if len(parts) > 1 && len(parts[0]) == 1 {
		return fmt.Sprintf("/cygdrive/%s/%s", strings.ToLower(parts[0]), strings.TrimPrefix(filepath.ToSlash(parts[1]), "/"))
	}
	return filepath.ToSlash(path)
}
