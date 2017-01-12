package kubectlapi

import (
	"github.com/continuouspipe/remote-environment-client/osapi"
)

func ConfigSetAuthInfo(environment string, username string, password string) string {
	nameParam := environment + "-" + username
	usernameFlag := "--username=" + username
	passwordFlag := "--password=" + password

	args := []string{kubeCtlName, "config", "set-credentials", nameParam, usernameFlag, passwordFlag}
	return osapi.CommandExec(appName, args...)
}

func ConfigSetCluster(environment string, clusterIp string) string {
	serverFlag := "--server=https://" + clusterIp

	args := []string{kubeCtlName, "config", "set-cluster", environment, serverFlag, "--insecure-skip-tls-verify=true"}
	return osapi.CommandExec(appName, args...)
}

func ConfigSetContext(environment string, username string) string {
	clusterFlag := "--cluster=" + environment
	userFlag := "--user=" + environment + "-" + username

	args := []string{kubeCtlName, "config", "set-context", environment, clusterFlag, userFlag}
	return osapi.CommandExec(appName, args...)
}
