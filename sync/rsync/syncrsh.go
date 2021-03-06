// +build !windows

package rsync

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"github.com/continuouspipe/remote-environment-client/sync/options"
	"github.com/continuouspipe/remote-environment-client/util/slice"
	"github.com/pkg/errors"
)

func init() {
	RsyncRsh = NewRSyncRsh()
}

type RSyncRsh struct {
	kubeConfigKey, environment, pod, remoteProjectPath string
	individualFileSyncThreshold                        int
	verbose, dryRun, delete                            bool
}

func NewRSyncRsh() *RSyncRsh {
	return &RSyncRsh{}
}

func (o *RSyncRsh) SetOptions(syncOptions options.SyncOptions) {
	o.kubeConfigKey = syncOptions.KubeConfigKey
	o.environment = syncOptions.Environment
	o.pod = syncOptions.Pod
	o.individualFileSyncThreshold = syncOptions.IndividualFileSyncThreshold
	o.remoteProjectPath = syncOptions.RemoteProjectPath
	o.verbose = syncOptions.Verbose
	o.dryRun = syncOptions.DryRun
	o.delete = syncOptions.Delete
}

func (o RSyncRsh) Sync(paths []string) error {
	cplogs.V(5).Infof("sync triggered for paths %s", paths)
	rsh := fmt.Sprintf(`%s %s --context=%s --namespace=%s exec -i %s`, config.AppName, config.KubeCtlName, o.kubeConfigKey, o.environment, o.pod)
	cplogs.V(5).Infof("setting RSYNC_RSH to %s\n", rsh)
	cplogs.Flush()
	os.Setenv("RSYNC_RSH", rsh)
	defer os.Unsetenv("RSYNC_RSH")

	args := []string{
		"-rlptDv",
		"--blocking-io",
		"--checksum",
		`--exclude=.git`}

	if o.delete {
		args = append(args, "--delete")
	}
	if o.verbose {
		args = append(args, "--verbose")
	}
	if o.dryRun {
		args = append(args, "--dry-run")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err := os.Stat(SyncFetchExcluded); err == nil {
		args = append(args, fmt.Sprintf(`--exclude-from=%s`, cwd+string(filepath.Separator)+SyncFetchExcluded))
	}

	paths = slice.RemoveDuplicateString(paths)

	paths, err = o.getRelativePathList(paths)
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "getting the current directory failed and is required for syncing").String())
	}

	allPathsExists, notExistingPaths := o.allPathsExists(paths)
	if !allPathsExists {
		cplogs.V(5).Infof("detected not existing path/s %s. We will do a generic rsync rather that an individual one", notExistingPaths)
		cplogs.Flush()
	}

	if len(paths) > 0 && len(paths) <= o.individualFileSyncThreshold && allPathsExists {
		cplogs.V(5).Infof("individual file sync, files to sync %d, threshold: %d", len(paths), o.individualFileSyncThreshold)
		cplogs.Flush()
		err = o.syncIndividualFiles(paths, args)
		if err != nil {
			return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "error when syncing individual files").String())
		}
	} else {
		err = o.syncAllFiles(paths, args)
		if err != nil {
			return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "error when syncing all files").String())
		}
	}
	return err
}

func (o RSyncRsh) allPathsExists(paths []string) (res bool, notExisting []string) {
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

func (o RSyncRsh) syncIndividualFiles(paths []string, args []string) error {
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
		lArgs = append(lArgs,
			"--include="+filepath.Base(path),
			"--exclude=*",
			"--",
			cwd+string(filepath.Separator)+filepath.Dir(path)+string(filepath.Separator),
			"--:"+o.remoteProjectPath+filepath.Dir(path)+string(filepath.Separator))

		err := o.executeRsync(lArgs, os.Stdout)
		if err != nil {
			errMsg := fmt.Sprintf("rsync failed to execute using arguments %s", lArgs)
			cplogs.V(4).Infof(errMsg)
			return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errMsg).String())
		}
	}

	return nil
}

func (o RSyncRsh) syncAllFiles(paths []string, args []string) error {
	args = append(args,
		"--relative",
		"--",
		"./",
		"--:"+o.remoteProjectPath,
	)
	err := o.executeRsync(args, os.Stdout)
	if err != nil {
		errMsg := fmt.Sprintf("rsync failed to execute using arguments %s", args)
		cplogs.V(4).Infof(errMsg)
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errMsg).String())
	}
	return err
}

func (rsh RSyncRsh) executeRsync(args []string, stdOut io.Writer) error {
	cplogs.V(5).Infof("rsync arguments: %s", args)
	scmd := osapi.SCommand{}
	scmd.Name = "rsync"
	scmd.Stdin = os.Stdin
	scmd.Stdout = stdOut
	scmd.Stderr = os.Stderr
	return osapi.CommandExecL(scmd, args...)
}

func (o RSyncRsh) getRelativePathList(paths []string) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "getting the current directory failed and is required for syncing").String())
	}
	for key, path := range paths {
		relPath, err := filepath.Rel(cwd, string(filepath.Separator)+path)
		if err != nil {
			return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, fmt.Sprintf("getting the relative path using cwd %s and path %s failed", cwd, path)).String())
		}
		paths[key] = relPath
	}

	return paths, nil
}
