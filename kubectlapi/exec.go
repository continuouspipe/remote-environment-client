package kubectlapi

import (
	"github.com/continuouspipe/remote-environment-client/osapi"
)

func CommandExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) string {
	args := getAllArgs(kubeConfigKey, environment, pod, execCmdArgs...)

	return osapi.CommandExec(appName, args...)
}

func SysCallExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) {
	args := getAllArgs(kubeConfigKey, environment, pod, execCmdArgs...)
	osapi.SysCallExec(appName, args...)
}

func getAllArgs(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) []string {
	contextFlag := "--context=" + kubeConfigKey
	namespaceFlag := "--namespace=" + environment
	kubeCmdArgs := []string{
		kubeCtlName, contextFlag, namespaceFlag, "exec", "-it", pod, "--",
	}
	allArgs := append(kubeCmdArgs, execCmdArgs...)
	return allArgs
}
