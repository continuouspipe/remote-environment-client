package sync

import (
	"fmt"
	"os"

	"github.com/continuouspipe/remote-environment-client/osapi"
	"github.com/continuouspipe/remote-environment-client/config"
)

/**
+         rsync -zrlptDv --blocking-io  --force --exclude=".*" --exclude-from="$(excludes_file)" \
+        -e 'kubectl --context='"$(context)"' --namespace='"$(namespace)"' exec -i '"$POD" -- --:/app/ "$(dir)"
+  */

func Fetch(kubeConfigKey string, environment string, service string, pod string) bool {
	currentDir, err := os.Getwd()
	if err != nil {
		return false
	}

	args := []string{
		config.KubeCtlName,
		"rsync",
		"-zrlptDv",
		"--blocking-io",
		"--force",
		`--exclude=".*"`,
		fmt.Sprintf(`--exclude-from="%s"`, SyncExcluded),
		"-e",
		fmt.Sprintf(`'%s --context='"%s"'`, config.KubeCtlName, kubeConfigKey),
		fmt.Sprintf(`--namespace='"%s"'`, environment),
		"exec",
		"-i",
		pod,
		"--",
		"--:/app/",
		currentDir,
	}

	fmt.Println(args)

	osapi.SysCallExec(config.AppName, args...)
	return true
}
