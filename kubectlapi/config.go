package kubectlapi

import (
	"os/exec"
)

const appName = "cp-remote-go"
const kubeCtlName = "kubectl"

func ConfigSetAuthInfo(namespace string, username string, password string) string {
	nameParam := namespace + "-" + username
	usernameFlag := "--username=" + username
	passwordFlag := "--password=" + password

	cmd := exec.Command(appName, kubeCtlName, "config", "set-credentials", nameParam, usernameFlag, passwordFlag)
	return executeCmd(cmd)
}

func ConfigSetCluster(namespace string, clusterIp string) string {
	serverFlag := "--server=https://" + clusterIp

	cmd := exec.Command(appName, kubeCtlName, "config", "set-cluster", namespace, serverFlag, "--insecure-skip-tls-verify=true")
	return executeCmd(cmd)
}

func ConfigSetContext(namespace string, username string) string {
	clusterFlag := "--cluster=" + namespace
	userFlag := "--user=" + namespace + "-" + username

	cmd := exec.Command(appName, kubeCtlName, "config", "set-context", namespace, clusterFlag, userFlag)
	return executeCmd(cmd)
}
