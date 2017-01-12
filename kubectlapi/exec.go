package kubectlapi

import "os/exec"

func Exec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) string {
	contextFlag := "--context=" + kubeConfigKey
	namespaceFlag := "--namespace=" + environment

	kubeCmdArgs := []string{
		kubeCtlName, contextFlag, namespaceFlag, "exec", "-it", pod, "--",
	}

	allArgs := append(kubeCmdArgs, execCmdArgs...)

	cmd := exec.Command(appName, allArgs...)
	return executeCmd(cmd)
}
