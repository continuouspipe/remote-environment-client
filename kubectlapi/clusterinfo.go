package kubectlapi

import (
	"github.com/continuouspipe/remote-environment-client/osapi"
)

func ClusterInfo(kubeConfigKey string) string {
	contextFlag := "--context=" + kubeConfigKey
	args := []string{kubeCtlName, contextFlag, "cluster-info"}

	return osapi.CommandExec(appName, args...)
}
