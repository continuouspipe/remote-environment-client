package kubectlapi

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/osapi"
)

// executes a command (spawn) on a specific pod
func CommandExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) string {
	args := getAllArgs(kubeConfigKey, environment, pod, execCmdArgs...)
	return osapi.CommandExec(config.AppName, args...)
}

// executes a system call (exec) on a specific pod
func SysCallExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) {
	args := getAllArgs(kubeConfigKey, environment, pod, execCmdArgs...)
	osapi.SysCallExec(config.AppName, args...)
}

// sets all the flags required to execute a command inside a container
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
