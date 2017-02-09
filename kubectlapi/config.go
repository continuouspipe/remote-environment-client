package kubectlapi

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/osapi"
)

func ConfigSetAuthInfo(environment string, username string, password string) (string, error) {
	args := []string{
		config.KubeCtlName,
		"config",
		"set-credentials",
		environment + "-" + username,
		"--username=" + username,
		"--password=" + password,
	}
	return osapi.CommandExec(getScmd(), args...)
}

func ConfigSetCluster(environment string, clusterIp string) (string, error) {
	args := []string{
		config.KubeCtlName,
		"config",
		"set-cluster",
		environment,
		"--server=https://" + clusterIp,
		"--insecure-skip-tls-verify=true",
	}
	return osapi.CommandExec(getScmd(), args...)
}

func ConfigSetContext(environment string, username string) (string, error) {
	args := []string{
		config.KubeCtlName,
		"config",
		"set-context",
		environment,
		"--cluster=" + environment,
		"--user=" + environment + "-" + username,
	}
	return osapi.CommandExec(getScmd(), args...)
}
