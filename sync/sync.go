package sync

import (
	"fmt"
	"os"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"k8s.io/client-go/pkg/api/v1"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"path/filepath"
)

const SyncExcluded = ".cp-remote-ignore"

type DirectoryEventSyncAll struct {
	KubeConfigKey, Environment  string
	Pod                         v1.Pod
	IndividualFileSyncThreshold int
}

func GetDirectoryEventSyncAll() *DirectoryEventSyncAll {
	return &DirectoryEventSyncAll{}
}

func (o *DirectoryEventSyncAll) OnLastChange(paths []string) error {
	rsh := fmt.Sprintf(`%s %s --context=%s --namespace=%s exec -i %s`, config.AppName, config.KubeCtlName, o.KubeConfigKey, o.Environment, o.Pod.GetName())
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

	if len(paths) > o.IndividualFileSyncThreshold {
		cplogs.V(5).Infof("batch file sync, files to sync %d, threshold: %d", len(paths), o.IndividualFileSyncThreshold)
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
		cplogs.V(5).Infof("individual file sync, files to sync %d, threshold: %d", len(paths), o.IndividualFileSyncThreshold)

		args = append(args, "--")
		args = append(args, relPaths...)
		args = append(args, "--:/app/")
	}

	cplogs.V(5).Infof("rsync arguments: %s", args)

	return osapi.CommandExecL("rsync", os.Stdout, args...)
}

func (o DirectoryEventSyncAll) getRelativePathList(paths []string) ([]string, error) {
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
