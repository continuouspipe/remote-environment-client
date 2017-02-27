package kubectlapi

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"strings"
)

type KubeCtlConfigProvider interface {
	ConfigSetAuthInfo(environment string, username string, password string) (string, error)
	ConfigSetCluster(environment string, clusterAddr string, teamName string, clusterIdentifier string) (string, error)
	ConfigSetContext(environment string, username string) (string, error)
}

type KubeCtlConfig struct{}

func NewKubeCtlConfig() *KubeCtlConfig {
	return &KubeCtlConfig{}
}

func (k KubeCtlConfig) ConfigSetAuthInfo(environment string, username string, password string) (string, error) {
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

func (k KubeCtlConfig) ConfigSetCluster(environment string, clusterAddr string, teamName string, clusterIdentifier string) (string, error) {

	//avoid double // in address
	clusterAddr = strings.TrimRight(clusterAddr, "/")

	args := []string{
		config.KubeCtlName,
		"config",
		"set-cluster",
		environment,
		fmt.Sprintf("--server=%s/%s/%s/", clusterAddr, teamName, clusterIdentifier),
		"--insecure-skip-tls-verify=true",
	}
	return osapi.CommandExec(getScmd(), args...)
}

func (k KubeCtlConfig) ConfigSetContext(environment string, username string) (string, error) {
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
