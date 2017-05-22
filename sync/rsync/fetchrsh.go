// +build !windows

package rsync

import (
	"fmt"
	"os"
	"path/filepath"

	"net/http"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"github.com/continuouspipe/remote-environment-client/sync/options"
	"github.com/pkg/errors"
)

func init() {
	RfetchRsh = NewRsyncRshFetch()
}

type RsyncRshFetch struct {
	kubeConfigKey, environment, pod, remoteProjectPath string
	verbose, dryRun                                    bool
}

func NewRsyncRshFetch() *RsyncRshFetch {
	return &RsyncRshFetch{}
}

func (r *RsyncRshFetch) SetOptions(syncOptions options.SyncOptions) {
	r.kubeConfigKey = syncOptions.KubeConfigKey
	r.environment = syncOptions.Environment
	r.pod = syncOptions.Pod
	r.remoteProjectPath = syncOptions.RemoteProjectPath
	r.verbose = syncOptions.Verbose
	r.dryRun = syncOptions.DryRun
}

func (r RsyncRshFetch) Fetch(filePath string) error {
	rsh := fmt.Sprintf(`%s %s --context=%s --namespace=%s exec -i %s`, config.AppName, config.KubeCtlName, r.kubeConfigKey, r.environment, r.pod)
	os.Setenv("RSYNC_RSH", rsh)
	defer os.Unsetenv("RSYNC_RSH")
	cplogs.V(5).Infof("setting RSYNC_RSH to %s\n", rsh)

	args := []string{
		"-zrlptDv",
		"--blocking-io",
		"--force",
		`--exclude=.git`,
	}

	if r.verbose {
		args = append(args, "--verbose")
	}
	if r.dryRun {
		args = append(args, "--dry-run")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "cannot fetch without knowing the cwd").String())
	}
	if _, err := os.Stat(FetchExcluded); err == nil {
		args = append(args, fmt.Sprintf(`--exclude-from=%s`, cwd+string(filepath.Separator)+FetchExcluded))
	}
	if _, err := os.Stat(SyncFetchExcluded); err == nil {
		args = append(args, fmt.Sprintf(`--exclude-from=%s`, cwd+string(filepath.Separator)+SyncFetchExcluded))
	}

	args = append(args, "--")

	if filePath == "" {
		cplogs.V(5).Infoln("fetching all files")
		args = append(args, "--:"+r.remoteProjectPath)
	} else {
		cplogs.V(5).Infof("fetching specified file %s", filePath)
		args = append(args, "--:"+r.remoteProjectPath+filePath)
	}

	args = append(args, cwd)

	cplogs.V(5).Infof("rsync arguments: %s", args)
	cplogs.Flush()

	scmd := osapi.SCommand{}
	scmd.Name = "rsync"
	scmd.Stdin = os.Stdin
	scmd.Stdout = os.Stdout
	scmd.Stderr = os.Stderr

	err = osapi.CommandExecL(scmd, args...)
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "executing rsync command has failed").String())
	}
	return err
}
