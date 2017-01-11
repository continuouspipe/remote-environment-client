package kubectlapi

import "os/exec"

func Exec(context string, namespace string, pod string, command string) string {
	contextFlag := "--context=" + context
	namespaceFlag := "--namespace=" + namespace
	cmd := exec.Command(appName, kubeCtlName, contextFlag, namespaceFlag, "exec", "-it", pod, "--", command)
	return executeCmd(cmd)
}
