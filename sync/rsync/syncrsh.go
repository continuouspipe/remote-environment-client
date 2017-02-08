// +build !windows

package rsync

import (
	"fmt"
	"os"
	"path/filepath"
	"io"
	"io/ioutil"

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

func (o RSyncRsh) Sync(paths []string) error {
	rsh := fmt.Sprintf(`%s %s --context=%s --namespace=%s exec -i %s`, config.AppName, config.KubeCtlName, o.kubeConfigKey, o.environment, o.pod)
	cplogs.V(5).Infof("setting RSYNC_RSH to %s\n", rsh)
	os.Setenv("RSYNC_RSH", rsh)
	defer os.Unsetenv("RSYNC_RSH")

	args := []string{
		"-rlptDv",
		"--delete",
		"--blocking-io",
		"--checksum"}

	var err error
	paths = slice.RemoveDuplicateString(paths)

	if len(paths) <= o.individualFileSyncThreshold {
		cplogs.V(5).Infof("individual file sync, files to sync %d, threshold: %d", len(paths), o.individualFileSyncThreshold)
		err = o.syncIndividualFiles(paths, args)
	} else {
		err = o.syncAllFiles(paths, args)
	}

	cplogs.Flush()
	return err
}

func (o RSyncRsh) syncIndividualFiles(paths []string, args []string) error {
	paths, err := o.getRelativePathList(paths)
	if err != nil {
		return err
	}

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
	errChan := make(chan error, 1)
	doneCountChan := make(chan bool, len(paths))

	for _, path := range paths {
		go func(lPath string, lArgs []string) {
			lArgs = append(lArgs,
				"--include="+filepath.Base(lPath),
				"--exclude=*",
				"--",
				cwd+"/"+filepath.Dir(lPath)+"/",
				"--:/app/"+filepath.Dir(lPath)+"/")

			fmt.Println(lPath)
			err := o.executeRsync(lArgs, ioutil.Discard)
			if err != nil {
				errChan <- err
			}
			doneCountChan <- true
		}(path, args)
	}

	doneCount := 0
	for {
		if doneCount == len(paths) {
			break
		}
		select {
		case <-doneCountChan:
			doneCount = doneCount + 1
		case err := <-errChan:
			return err
		}
	}

	return nil
}

func (o RSyncRsh) syncAllFiles(paths []string, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err := os.Stat(SyncExcluded); err == nil {
		args = append(args, fmt.Sprintf(`--exclude-from=%s`, cwd+"/"+SyncExcluded))
	}
	args = append(args,
		"--relative",
		"--",
		"./",
		"--:/app/",
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
		relPath, err := filepath.Rel(cwd, "/"+path)
		if err != nil {
			return nil, err
		}
		paths[key] = relPath
	}

	return paths, nil
}
