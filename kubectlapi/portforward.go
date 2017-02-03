package kubectlapi

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/osapi"
)

func Forward(kubeConfigKey string, environment string, pod string, ports string, killProcess chan bool) error {
	args := []string{
		config.KubeCtlName,
		"--context=" + kubeConfigKey,
		"--namespace=" + environment,
		"port-forward",
		pod,
		ports,
	}

	return osapi.StartProcess(getScmd(), killProcess, args...)
}
