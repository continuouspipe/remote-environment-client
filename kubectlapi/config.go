package kubectlapi

import (
	"os/exec"
)

func ConfigSetAuthInfo(environment string, username string, password string) string {
	nameParam := environment + "-" + username
	usernameFlag := "--username=" + username
	passwordFlag := "--password=" + password

	cmd := exec.Command(appName, kubeCtlName, "config", "set-credentials", nameParam, usernameFlag, passwordFlag)
	return executeCmd(cmd)
}

func ConfigSetCluster(environment string, clusterIp string) string {
	serverFlag := "--server=https://" + clusterIp

	cmd := exec.Command(appName, kubeCtlName, "config", "set-cluster", environment, serverFlag, "--insecure-skip-tls-verify=true")
	return executeCmd(cmd)
}

func ConfigSetContext(environment string, username string) string {
	clusterFlag := "--cluster=" + environment
	userFlag := "--user=" + environment + "-" + username

	cmd := exec.Command(appName, kubeCtlName, "config", "set-context", environment, clusterFlag, userFlag)
	return executeCmd(cmd)
}
