package kubectlapi

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"os"
)

type KubeCtlClusterInfoProvider interface {
	ClusterInfo(kubeConfigKey string) (string, error)
}

type KubeCtlClusterInfo struct{}

func NewKubeCtlClusterInfo() *KubeCtlClusterInfo {
	return &KubeCtlClusterInfo{}
}

func (k KubeCtlClusterInfo) ClusterInfo(kubeConfigKey string) (string, error) {
	args := []string{
		config.KubeCtlName,
		"--context=" + kubeConfigKey,
		"cluster-info"}
	return osapi.CommandExec(getScmd(), args...)
}

func getScmd() osapi.SCommand {
	scmd := osapi.SCommand{}
	scmd.Name = config.AppName
	scmd.Stdin = os.Stdin
	scmd.Stdout = os.Stdout
	scmd.Stderr = os.Stderr
	return scmd
}
