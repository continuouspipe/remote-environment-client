package kubectlapi

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/osapi"
)

func Forward(kubeConfigKey string, environment string, pod string, ports string) error {
	args := []string{
		config.KubeCtlName,
		"--context=" + kubeConfigKey,
		"--namespace=" + environment,
		"port-forward",
		pod,
		ports,
	}

	return osapi.StartProcess(config.AppName, args...)
}
