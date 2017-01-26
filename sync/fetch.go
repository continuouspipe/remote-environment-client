package sync

import (
	"fmt"
	"os"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/osapi"
)

type Fetcher interface {
	Fetch(kubeConfigKey string, environment string, pod string) error
}

type RsyncFetch struct{}

func NewRsyncFetch() *RsyncFetch {
	return &RsyncFetch{}
}

func (r RsyncFetch) Fetch(kubeConfigKey string, environment string, pod string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
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

	return osapi.CommandExecL("rsync", os.Stdout, args...)
}
