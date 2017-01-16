package sync

import (
	"fmt"
	"os"

	"github.com/continuouspipe/remote-environment-client/osapi"
	"github.com/continuouspipe/remote-environment-client/config"
)

type Fetcher interface {
	Fetch(kubeConfigKey string, environment string, pod string) bool
}

type RsyncFetch struct{}

func NewRsyncFetch() *RsyncFetch {
	return &RsyncFetch{}
}

func (r RsyncFetch) Fetch(kubeConfigKey string, environment string, pod string) bool {
	currentDir, err := os.Getwd()
	if err != nil {
		return false
	}

	os.Setenv("RSYNC_RSH", fmt.Sprintf(`%s %s --context=%s --namespace=%s exec -i %s`, config.AppName, config.KubeCtlName, kubeConfigKey, environment, pod))
	defer os.Unsetenv("RSYNC_RSH")

	args := []string{
		"-zrlptDv",
		"--blocking-io",
		"--force",
		`--exclude=".*"`,
		fmt.Sprintf(`--exclude-from=%s`, SyncExcluded),
		"--",
		"--:/app/",
		currentDir,
	}

	osapi.SysCallExec("rsync", args...)
	return true
}
