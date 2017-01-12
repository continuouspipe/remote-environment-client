package kubectlapi

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/osapi"
)

func CommandExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) string {
	args := getAllArgs(kubeConfigKey, environment, pod, execCmdArgs...)
	return osapi.CommandExec(config.AppName, args...)
}

func SysCallExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) {
	args := getAllArgs(kubeConfigKey, environment, pod, execCmdArgs...)
	osapi.SysCallExec(config.AppName, args...)
}

func getAllArgs(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) []string {
	kubeCmdArgs := []string{
		config.KubeCtlName,
		"--context=" + kubeConfigKey,
		"--namespace=" + environment,
		"exec",
		"-it",
		pod,
		"--",
	}

	allArgs := append(kubeCmdArgs, execCmdArgs...)
	return allArgs
}
