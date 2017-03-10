// +build !windows

package rsync

import (
	"fmt"
	"os"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/osapi"
)

func init() {
	RfetchRsh = NewRsyncRshFetch()
}

type RsyncRshFetch struct {
	kubeConfigKey, environment string
	pod                        string
	remoteProjectPath          string
}

func NewRsyncRshFetch() *RsyncRshFetch {
	return &RsyncRshFetch{}
}

func (r *RsyncRshFetch) SetKubeConfigKey(kubeConfigKey string) {
	r.kubeConfigKey = kubeConfigKey
}

func (r *RsyncRshFetch) SetEnvironment(environment string) {
	r.environment = environment
}

func (r *RsyncRshFetch) SetPod(pod string) {
	r.pod = pod
}

func (r *RsyncRshFetch) SetRemoteProjectPath(remoteProjectPath string) {
	r.remoteProjectPath = remoteProjectPath
}

func (r RsyncRshFetch) Fetch(filePath string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	rsh := fmt.Sprintf(`%s %s --context=%s --namespace=%s exec -i %s`, config.AppName, config.KubeCtlName, r.kubeConfigKey, r.environment, r.pod)
	os.Setenv("RSYNC_RSH", rsh)
	defer os.Unsetenv("RSYNC_RSH")
	cplogs.V(5).Infof("setting RSYNC_RSH to %s\n", rsh)

	args := []string{
		"-zrlptDv",
		"--blocking-io",
		"--force",
		`--exclude=.git`,
		fmt.Sprintf(`--exclude-from=%s`, SyncExcluded),
		"--",
	}

	if filePath == "" {
		cplogs.V(5).Infoln("fetching all files")
		args = append(args, "--:"+r.remoteProjectPath)
	} else {
		cplogs.V(5).Infof("fetching specified file %s", filePath)
		args = append(args, "--:"+r.remoteProjectPath+filePath)
	}

	args = append(args, currentDir)

	cplogs.V(5).Infof("rsync arguments: %s", args)
	cplogs.Flush()

	scmd := osapi.SCommand{}
	scmd.Name = "rsync"
	scmd.Stdin = os.Stdin
	scmd.Stdout = os.Stdout
	scmd.Stderr = os.Stderr

	return osapi.CommandExecL(scmd, args...)
}
