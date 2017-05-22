// +build windows
package rsync

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/continuouspipe/remote-environment-client/cplogs"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	kexec "github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"github.com/continuouspipe/remote-environment-client/sync/options"
	"github.com/continuouspipe/remote-environment-client/util/slice"
	"github.com/pkg/errors"
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
		errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "start daemon on random port failed").String())
	}
	defer r.remoteRsync.KillDaemon(pidFile)

	stopChan, err := r.remoteRsync.StartPortForwardOnRandomPort()
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "start port forward on random port failed").String())
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
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "getting the current directory failed and is required for syncing").String())
	}
	//check if sync fetch excluded file exists, if it doesn't, don't return an error
	if _, err := os.Stat(SyncFetchExcluded); err == nil {
		args = append(args, fmt.Sprintf(`--exclude-from=%s`, cwd+"/"+SyncFetchExcluded))
	}

	paths = slice.RemoveDuplicateString(paths)

	paths, err = r.getRelativePathList(paths)
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "error when gettin the relative path list").String())
	}

	allPathsExists, notExistingPaths := r.allPathsExists(paths)
	if !allPathsExists {
		cplogs.V(5).Infof("detected not existing path/s %s. We will do a generic rsync rather that an individual one", notExistingPaths)
		cplogs.Flush()
	}

	if len(paths) > 0 && len(paths) <= r.individualFileSyncThreshold && allPathsExists {
		cplogs.V(5).Infof("individual file sync, files to sync %d, threshold: %d", len(paths), r.individualFileSyncThreshold)
		cplogs.Flush()
		err = r.syncIndividualFiles(paths, args)
		if err != nil {
			return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "error when syncing individual files").String())
		}
	} else {
		err = r.syncAllFiles(paths, args)
		if err != nil {
			return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "error when syncing all files").String())
		}
	}

	return nil
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
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "getting the current directory failed and is required for syncing").String())
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
			errMsg := fmt.Sprintf("rsync failed to execute using arguments %s", lArgs)
			cplogs.V(4).Infof(errMsg)
			return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errMsg).String())
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
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "getting the current directory failed and is required for syncing").String())
	}
	for key, path := range paths {
		relPath, err := filepath.Rel(cwd, path)
		if err != nil {
			return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, fmt.Sprintf("getting the relative path using cwd %s and path %s failed", cwd, path)).String())
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
