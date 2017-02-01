package sync

import (
	"fmt"
	"os"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"k8s.io/client-go/pkg/api/v1"
	"github.com/continuouspipe/remote-environment-client/osapi"
)

const SyncExcluded = ".cp-remote-ignore"

type DirectoryEventSyncAll struct {
	KubeConfigKey, Environment string
	Pod                        v1.Pod
}

func GetDirectoryEventSyncAll() *DirectoryEventSyncAll {
	return &DirectoryEventSyncAll{}
}

func (o *DirectoryEventSyncAll) OnLastChange() error {
	rsh := fmt.Sprintf(`%s %s --context=%s --namespace=%s exec -i %s`, config.AppName, config.KubeCtlName, o.KubeConfigKey, o.Environment, o.Pod.GetName())
	cplogs.V(5).Infof("setting RSYNC_RSH to %s\n", rsh)
	os.Setenv("RSYNC_RSH", rsh)
	defer os.Unsetenv("RSYNC_RSH")

	args := []string{
		"-rlptDv",
		"--delete",
		"--relative",
		"--blocking-io"}

	if _, err := os.Stat(SyncExcluded); err == nil {
		args = append(args, fmt.Sprintf(`--exclude-from=%s`, SyncExcluded))
	}

	args = append(args,
		"--",
		"./",
		"--:/app/",
	)

	cplogs.V(5).Infof("rsync arguments: %s", args)

	return osapi.CommandExecL("rsync", os.Stdout, args...)
}
