package kubectlapi

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/osapi"
)

func ClusterInfo(kubeConfigKey string) string {
	args := []string{
		config.KubeCtlName,
		"--context=" + kubeConfigKey,
		"cluster-info"}

	return osapi.CommandExec(config.AppName, args...)
}
