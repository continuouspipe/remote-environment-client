package kubectlapi

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/osapi"
)

func Forward(pod string, ports string) {
	args := []string{
		config.KubeCtlName,
		"port-forward",
		pod,
		ports,
	}

	osapi.SysCallExec(config.AppName, args...)
}
