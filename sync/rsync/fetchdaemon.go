// +build windows
package rsync

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	kexec "github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/osapi"
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
	commandTimeout       = 10
	differentPortTimeout = 20
	rsyncConfigSection   = "root"
	configFile           = "rsyncd.conf"
	pidFile              = "rsyncd-pid.conf"
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
	remoteRsync                *RemoteRsyncDeamon
	kubeConfigKey, environment string
	pod                        string
	remoteProjectPath          string
}

func (r *RsyncDaemonFetch) SetKubeConfigKey(kubeConfigKey string) {
	r.kubeConfigKey = kubeConfigKey
}

func (r *RsyncDaemonFetch) SetEnvironment(environment string) {
	r.environment = environment
}

func (r *RsyncDaemonFetch) SetPod(pod string) {
	r.pod = pod
}

func (r *RsyncDaemonFetch) SetRemoteProjectPath(remoteProjectPath string) {
	r.remoteProjectPath = remoteProjectPath
}

func (r RsyncDaemonFetch) Fetch(filePath string) error {
	kscmd := kexec.KSCommand{}
	kscmd.KubeConfigKey = r.kubeConfigKey
	kscmd.Environment = r.environment
	kscmd.Pod = r.pod
	kscmd.Stderr = ioutil.Discard
	kscmd.Stdout = ioutil.Discard
	r.remoteRsync.SetKSCommand(kscmd)

	err := r.remoteRsync.StartDaemonOnRandomPort()
	if err != nil {
		return err
	}
	defer r.remoteRsync.KillDaemon(pidFile)

	stopChan, err := r.remoteRsync.StartPortForwardOnRandomPort()
	if err != nil {
		return err
	}
	defer r.remoteRsync.StopPortForward(stopChan)

	args := []string{
		"-zrlDv",
		"--omit-dir-times",
		"--blocking-io",
		"--force",
		`--exclude=.*`,
		fmt.Sprintf(`--exclude-from=%s`, SyncExcluded),
		"--",
	}

	if filePath == "" {
		cplogs.V(5).Infoln("fetching all files")
		args = append(args, r.remoteRsync.GetRsyncURL(rsyncConfigSection, r.remoteProjectPath))
	} else {
		cplogs.V(5).Infof("fetching specified file %s", filePath)
		args = append(args, r.remoteRsync.GetRsyncURL(rsyncConfigSection, r.remoteProjectPath+filePath))
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
	executor      kexec.Executor
	kscmd         kexec.KSCommand
	localPort     int
	remotePort    int
	portRangeFrom int
	portRangeTo   int
}

func NewRemoteRsyncDeamon() *RemoteRsyncDeamon {
	r := &RemoteRsyncDeamon{}
	r.portRangeFrom = 30000
	r.portRangeTo = 40000
	r.executor = kexec.NewLocal()
	return r
}

func (r *RemoteRsyncDeamon) SetKSCommand(kscmd kexec.KSCommand) {
	r.kscmd = kscmd
}

// sets the remote port number to a random one between the range specified
// it then start the rsync daemon and waits for it to start
// if an error occurs it tries with a different port
// the attempts continues until a timeout value is reached
func (r *RemoteRsyncDeamon) StartDaemonOnRandomPort() error {
	startTime := time.Now()
	for {
		r.setRandomRemotePort()
		errChan := r.startDaemon(configFile, pidFile)
		err := r.waitForDaemon(pidFile, errChan)
		if err == nil {
			break
		}
		if time.Since(startTime) > differentPortTimeout*time.Second {
			return err
		}
	}
	return nil
}

//run the remote script that kills the rsync daemon
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

// sets the local port number to a random one between the range specified
// it then attempt to establish port forwarding from the local port to the remote one
// if an error occurs it tires with a different port
// the attempts continues until a timeout value is reached
func (r *RemoteRsyncDeamon) StartPortForwardOnRandomPort() (*chan bool, error) {
	startTime := time.Now()
	var stopChan *chan bool
	var err error
	for {
		r.setRandomLocalPort()
		stopChan, err = r.startPortForward()
		if err == nil {
			break
		}
		if time.Since(startTime) > differentPortTimeout*time.Second {
			return nil, err
		}
	}
	return stopChan, nil
}

// closes the channel that will then kill the goroutine that is running the port forwarding
func (r RemoteRsyncDeamon) StopPortForward(stopChan *chan bool) {
	if stopChan == nil {
		cplogs.V(5).Infoln("was not possible to stop the port forwarding")
		return
	}
	cplogs.V(5).Infoln("stopping port forwarding")
	close(*stopChan)
}

func (r RemoteRsyncDeamon) GetRsyncURL(label string, path string) string {
	return fmt.Sprintf("rsync://127.0.0.1:%d/%s/%s", r.localPort, label, strings.TrimPrefix(path, "/"))
}

func (r RemoteRsyncDeamon) startDaemon(configFile string, pidFile string) *chan error {
	start := fmt.Sprintf(startDaemon, configFile, pidFile, r.remotePort)
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

func (r RemoteRsyncDeamon) waitForDaemon(pidFile string, errChan *chan error) error {
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

func (r RemoteRsyncDeamon) startPortForward() (*chan bool, error) {

	sLocalPort := strconv.Itoa(r.localPort)
	sRemotePort := strconv.Itoa(r.remotePort)

	killProcess := make(chan bool, 1)
	cplogs.V(5).Infoln("starting port forward in the background")
	go func() {
		err := kubectlapi.Forward(r.kscmd.KubeConfigKey, r.kscmd.Environment, r.kscmd.Pod, sLocalPort+":"+sRemotePort, killProcess)
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
		cplogs.V(5).Infof("verifying if %s is open", "127.0.0.1:"+sLocalPort)
		conn, err := net.DialTimeout("tcp", "127.0.0.1:"+sLocalPort, 2*time.Second)
		if err == nil {
			cplogs.V(5).Infof("%s is open", "127.0.0.1:"+sLocalPort)
			conn.Close()
			break
		}
		cplogs.V(5).Infof("%s is still closed, retrying..", "127.0.0.1:"+sLocalPort)
		time.Sleep(100 * time.Millisecond)
	}
	return &killProcess, nil
}

func (r *RemoteRsyncDeamon) setRandomLocalPort() {
	r.localPort = r.getRandomPort()
}
func (r *RemoteRsyncDeamon) setRandomRemotePort() {
	r.remotePort = r.getRandomPort()
}

func (r RemoteRsyncDeamon) getRandomPort() int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(r.portRangeTo-r.portRangeFrom) + r.portRangeFrom
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
