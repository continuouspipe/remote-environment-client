package rsync

import (
	"fmt"
	"runtime"
	"os"
	"io/ioutil"
	"path/filepath"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"github.com/continuouspipe/remote-environment-client/util/slice"
	kexec "github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
)

func init() {
	RsyncDaemon = NewRSyncDaemon()
}

type RSyncDaemon struct {
	kubeConfigKey, environment  string
	pod                         string
	individualFileSyncThreshold int
	remoteRsync                 *RemoteRsyncDeamon
}

func NewRSyncDaemon() *RSyncDaemon {
	d := &RSyncDaemon{}
	d.remoteRsync = NewRemoteRsyncDeamon()
	return d
}

func (r *RSyncDaemon) SetKubeConfigKey(kubeConfigKey string) {
	r.kubeConfigKey = kubeConfigKey
}
func (r *RSyncDaemon) SetEnvironment(environment string) {
	r.environment = environment
}
func (r *RSyncDaemon) SetPod(pod string) {
	r.pod = pod
}
func (r *RSyncDaemon) SetIndividualFileSyncThreshold(individualFileSyncThreshold int) {
	r.individualFileSyncThreshold = individualFileSyncThreshold
}

func (r *RSyncDaemon) Sync(filePaths []string) error {
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
		"-rlptDv",
		"--delete",
		"--relative",
		"--blocking-io",
		"--checksum"}

	if _, err := os.Stat(SyncExcluded); err == nil {
		args = append(args, fmt.Sprintf(`--exclude-from=%s`, SyncExcluded))
	}

	remoteRsyncUrl := r.remoteRsync.GetRsyncURL(rsyncConfigSection, "app")

	filePaths = slice.RemoveDuplicateString(filePaths)

	if len(filePaths) > r.individualFileSyncThreshold {
		cplogs.V(5).Infof("batch file sync, files to sync %d, threshold: %d", len(filePaths), r.individualFileSyncThreshold)
		args = append(args,
			"--",
			".",
		)
		args = append(args, remoteRsyncUrl)
	} else {
		relPaths, err := r.getRelativePathList(filePaths)
		if err != nil {
			return err
		}
		cplogs.V(5).Infof("individual file sync, files to sync %d, threshold: %d", len(filePaths), r.individualFileSyncThreshold)

		args = append(args, "--")
		args = append(args, relPaths...)
		args = append(args, remoteRsyncUrl)
	}

	cplogs.V(5).Infof("rsync arguments: %s", args)
	cplogs.Flush()

	scmd := osapi.SCommand{}
	scmd.Name = "rsync"
	scmd.Stdin = os.Stdin
	scmd.Stdout = os.Stdout
	scmd.Stderr = os.Stderr

	return osapi.CommandExecL(scmd, args...)
}

func (o RSyncDaemon) getRelativePathList(paths []string) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	for key, path := range paths {
		relPath, err := filepath.Rel(cwd, path)
		if err != nil {
			return nil, err
		}

		if runtime.GOOS == "windows" {
			paths[key] = convertWindowsPath(relPath)
		} else {
			paths[key] = relPath
		}
	}

	return paths, nil
}
