package kubectlapi

import "os/exec"

func ClusterInfo(context string) string {
	contextFlag := "--context=" + context
	cmd := exec.Command(appName, kubeCtlName, contextFlag, "cluster-info")
	return executeCmd(cmd)
}
