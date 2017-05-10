// +build windows
package rsync

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	kexec "github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"github.com/continuouspipe/remote-environment-client/sync/options"
	"github.com/continuouspipe/remote-environment-client/util/slice"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

func init() {
	RsyncDaemon = NewRSyncDaemon()
}

type RSyncDaemon struct {
	kubeConfigKey, environment, pod, remoteProjectPath string
	individualFileSyncThreshold                        int
	remoteRsync                                        *RemoteRsyncDeamon
	verbose, dryRun, delete                            bool
}

func NewRSyncDaemon() *RSyncDaemon {
	d := &RSyncDaemon{}
	d.remoteRsync = NewRemoteRsyncDeamon()
	return d
}

func (r *RSyncDaemon) SetOptions(syncOptions options.SyncOptions) {
	r.kubeConfigKey = syncOptions.KubeConfigKey
	r.environment = syncOptions.Environment
	r.pod = syncOptions.Pod
	r.individualFileSyncThreshold = syncOptions.IndividualFileSyncThreshold
	r.remoteProjectPath = syncOptions.RemoteProjectPath
	r.verbose = syncOptions.Verbose
	r.dryRun = syncOptions.DryRun
	r.delete = syncOptions.Delete
}

func (r *RSyncDaemon) Sync(paths []string) error {
	kscmd := kexec.KSCommand{}
	kscmd.KubeConfigKey = r.kubeConfigKey
	kscmd.Environment = r.environment
	kscmd.Pod = r.pod
	kscmd.Stderr = ioutil.Discard
	kscmd.Stdout = ioutil.Discard
	r.remoteRsync.SetKSCommand(kscmd)

	err := r.remoteRsync.StartDaemonOnRandomPort()
	if err != nil {
		//TODO: Wrap the error making it Stateful

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
		"--checksum",
		`--exclude=.git`}

	if r.delete {
		args = append(args, "--delete")
	}
	if r.verbose {
		args = append(args, "--verbose")
	}
	if r.dryRun {
		args = append(args, "--dry-run")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err := os.Stat(SyncFetchExcluded); err == nil {
		args = append(args, fmt.Sprintf(`--exclude-from=%s`, cwd+"/"+SyncFetchExcluded))
	}

	paths = slice.RemoveDuplicateString(paths)

	paths, err = r.getRelativePathList(paths)
	if err != nil {
		return err
	}

	allPathsExists, notExistingPaths := r.allPathsExists(paths)
	if !allPathsExists {
		cplogs.V(5).Infof("detected not existing path/s %s. We will do a generic rsync rather that an individual one", notExistingPaths)
	}

	if len(paths) > 0 && len(paths) <= r.individualFileSyncThreshold && allPathsExists {
		cplogs.V(5).Infof("individual file sync, files to sync %d, threshold: %d", len(paths), r.individualFileSyncThreshold)
		err = r.syncIndividualFiles(paths, args)
	} else {
		err = r.syncAllFiles(paths, args)
	}

	cplogs.Flush()
	return err
}

func (o RSyncDaemon) allPathsExists(paths []string) (res bool, notExisting []string) {
	for _, path := range paths {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			notExisting = append(notExisting, path)
		} else if err != nil {
			return false, []string{path}
		}
	}
	return len(notExisting) == 0, notExisting
}

func (o RSyncDaemon) syncIndividualFiles(paths []string, args []string) error {
	remoteRsyncUrl := o.remoteRsync.GetRsyncURL(rsyncConfigSection, o.remoteProjectPath)

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	//this is a workaround to the issue with --delete throwing an error if the local file has been deleted
	//which we want to delete in the remote pod.
	//e.g( rsync: link_stat "/path/to/file.txt" failed: No such file or directory (2))
	//
	//as a fix, for each path we need to run a separate rsync command as we need to specify the relative path of the directory
	//in the target. this is necessary because --include only works for files in the first level of the target directory
	//and using --include is the only way to be able to delete a file remotely that doesn't exist locally
	//and prevents the "rsync: link_stat" error above

	for _, path := range paths {
		lArgs := args
		baseDir := cwd + string(filepath.Separator) + filepath.Dir(path) + string(filepath.Separator)
		lArgs = append(args,
			"--include="+filepath.Base(path),
			"--exclude=*",
			"--",
			convertWindowsPath(baseDir),
			remoteRsyncUrl+filepath.Dir(path)+"/")

		fmt.Println(path)
		err := o.executeRsync(lArgs, ioutil.Discard)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o RSyncDaemon) syncAllFiles(paths []string, args []string) error {
	remoteRsyncUrl := o.remoteRsync.GetRsyncURL(rsyncConfigSection, o.remoteProjectPath)
	args = append(args,
		"--relative",
		"--",
		".",
		remoteRsyncUrl,
	)
	return o.executeRsync(args, os.Stdout)
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

func (rsh RSyncDaemon) executeRsync(args []string, stdOut io.Writer) error {
	cplogs.V(5).Infof("rsync arguments: %s", args)
	scmd := osapi.SCommand{}
	scmd.Name = "rsync"
	scmd.Stdin = os.Stdin
	scmd.Stdout = stdOut
	scmd.Stderr = os.Stderr
	return osapi.CommandExecL(scmd, args...)
}
