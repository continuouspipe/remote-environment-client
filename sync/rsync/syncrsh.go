// +build !windows

package rsync

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/osapi"
)

func init() {
	Rsync = NewRSyncRsh()
}

const SyncExcluded = ".cp-remote-ignore"

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

func (o *RSyncRsh) Sync(paths []string) error {
	rsh := fmt.Sprintf(`%s %s --context=%s --namespace=%s exec -i %s`, config.AppName, config.KubeCtlName, o.kubeConfigKey, o.environment, o.pod)
	cplogs.V(5).Infof("setting RSYNC_RSH to %s\n", rsh)
	os.Setenv("RSYNC_RSH", rsh)
	defer os.Unsetenv("RSYNC_RSH")

	args := []string{
		"-rlptDv",
		"--delete",
		"--relative",
		"--blocking-io",
		"--checksum"}

	if _, err := os.Stat(SyncExcluded); err == nil {
		args = append(args, fmt.Sprintf(`--exclude-from=%s`, SyncExcluded))
	}

	if len(paths) > o.individualFileSyncThreshold {
		cplogs.V(5).Infof("batch file sync, files to sync %d, threshold: %d", len(paths), o.individualFileSyncThreshold)
		args = append(args,
			"--",
			"./",
			"--:/app/",
		)
	} else {
		relPaths, err := o.getRelativePathList(paths)
		if err != nil {
			return err
		}
		cplogs.V(5).Infof("individual file sync, files to sync %d, threshold: %d", len(paths), o.individualFileSyncThreshold)

		args = append(args, "--")
		args = append(args, relPaths...)
		args = append(args, "--:/app/")
	}

	cplogs.V(5).Infof("rsync arguments: %s", args)

	return osapi.CommandExecL("rsync", os.Stdout, args...)
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
