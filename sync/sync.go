package sync

import (
	"os"
	"fmt"
	"os/exec"
	"bufio"

	"k8s.io/client-go/pkg/api/v1"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
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
		"-av",
		"--delete",
		"--relative",
		"--blocking-io",
		fmt.Sprintf(`--exclude-from=%s`, SyncExcluded),
		"--",
		"./",
		"--:/app/",
	}

	cplogs.V(5).Infof("rsync arguments: %s", args)

	cmd := exec.Command("rsync", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cplogs.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		cplogs.Fatal(err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		cplogs.V(5).Infoln(line)
	}

	return nil
}
