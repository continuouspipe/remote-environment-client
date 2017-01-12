package kubectlapi

import (
	"os/exec"
	"syscall"
	"os"
)

func CommandExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) string {
	allArgs := getAllArgs(kubeConfigKey, environment, pod, execCmdArgs...)

	cmd := exec.Command(appName, allArgs...)
	return executeCmd(cmd)
}

func SysCallExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) {
	allArgs := getAllArgs(kubeConfigKey, environment, pod, execCmdArgs...)

	appBinPath, lookErr := exec.LookPath(appName)
	if lookErr != nil {
		panic(lookErr)
	}

	env := os.Environ()

	//syscall.Exec requires the first argument to be the app-name
	allArgs = append([]string{appName}, allArgs...)

	execErr := syscall.Exec(appBinPath, allArgs, env)
	if execErr != nil {
		panic(execErr)
	}
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
