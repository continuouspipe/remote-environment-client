// +build !windows

package rsync

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"github.com/continuouspipe/remote-environment-client/util/slice"
)

func init() {
	RsyncRsh = NewRSyncRsh()
}

type RSyncRsh struct {
	kubeConfigKey, environment  string
	pod                         string
	individualFileSyncThreshold int
	remoteProjectPath           string
}

func NewRSyncRsh() *RSyncRsh {
	return &RSyncRsh{}
}

func (o *RSyncRsh) SetKubeConfigKey(kubeConfigKey string) {
	o.kubeConfigKey = kubeConfigKey
}

func (o *RSyncRsh) SetEnvironment(environment string) {
	o.environment = environment
}

func (o *RSyncRsh) SetPod(pod string) {
	o.pod = pod
}

func (o *RSyncRsh) SetIndividualFileSyncThreshold(individualFileSyncThreshold int) {
	o.individualFileSyncThreshold = individualFileSyncThreshold
}

func (o *RSyncRsh) SetRemoteProjectPath(remoteProjectPath string) {
	o.remoteProjectPath = remoteProjectPath
}

func (o RSyncRsh) Sync(paths []string) error {
	rsh := fmt.Sprintf(`%s %s --context=%s --namespace=%s exec -i %s`, config.AppName, config.KubeCtlName, o.kubeConfigKey, o.environment, o.pod)
	cplogs.V(5).Infof("setting RSYNC_RSH to %s\n", rsh)
	os.Setenv("RSYNC_RSH", rsh)
	defer os.Unsetenv("RSYNC_RSH")

	args := []string{
		"-rlptDv",
		"--delete",
		"--blocking-io",
		"--checksum",
		`--exclude=.git`}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err := os.Stat(SyncExcluded); err == nil {
		args = append(args, fmt.Sprintf(`--exclude-from=%s`, cwd+string(filepath.Separator)+SyncExcluded))
	}

	paths = slice.RemoveDuplicateString(paths)

	paths, err = o.getRelativePathList(paths)
	if err != nil {
		return err
	}

	allPathsExists, notExistingPaths := o.allPathsExists(paths)
	if !allPathsExists {
		cplogs.V(5).Infof("detected not existing path/s %s. We will do a generic rsync rather that an individual one", notExistingPaths)
	}

	if len(paths) > 0 && len(paths) <= o.individualFileSyncThreshold && allPathsExists {
		cplogs.V(5).Infof("individual file sync, files to sync %d, threshold: %d", len(paths), o.individualFileSyncThreshold)
		err = o.syncIndividualFiles(paths, args)
	} else {
		err = o.syncAllFiles(paths, args)
	}

	cplogs.Flush()
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
			return err
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
	return o.executeRsync(args, os.Stdout)
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
		return nil, err
	}
	for key, path := range paths {
		relPath, err := filepath.Rel(cwd, string(filepath.Separator)+path)
		if err != nil {
			return nil, err
		}
		paths[key] = relPath
	}

	return paths, nil
}
