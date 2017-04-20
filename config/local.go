package config

import (
	"github.com/spf13/viper"
)

type localConfig struct {
	viperWrapper
}

const (
	InitToken           = "init-token"
	FlowId              = "flow-id"
	ClusterIdentifier   = "cluster-identifier"
	KubeEnvironmentName = "kube-environment-name"
	RemoteName          = "remote-name"
	RemoteBranch        = "remote-branch"
	Service             = "service"
	AnybarPort          = "anybar-port"
	KeenWriteKey        = "keen-write-key"
	KeenProjectId       = "keen-project-id"
	KeenEventCollection = "keen-event-collection"
	RemoteEnvironmentId = "remote-environment-id"
	InitStatus          = "init-status"

	//settings to disable the kube proxy if required
	CpKubeProxyEnabled        = "kube-proxy-enabled"
	KubeDirectClusterAddr     = "kube-direct-cluster-addr"
	KubeDirectClusterUser     = "kube-direct-cluster-user"
	KubeDirectClusterPassword = "kube-direct-cluster-password"
)

func newLocalConfig() *localConfig {
	local := &localConfig{}
	local.settings = []Setting{
		{InitToken, "", true},                  //Token used when the project was initialised
		{FlowId, "", true},                     //CP flow uuid
		{ClusterIdentifier, "", true},          //CP cluster Identifier
		{KubeEnvironmentName, "", true},        //Kubernetes environment name
		{RemoteName, "origin", true},           //Github remote name (origin by default)
		{RemoteBranch, "", true},               //Git name of the git branch used for the remote environment
		{Service, "web", true},                 //Kubernetes service name for the commands like (watch, bash, fetch and resync)
		{AnybarPort, "", false},                //AnyBar port number
		{KeenWriteKey, "", false},              //Keen.io write key
		{KeenProjectId, "", false},             //Keen.io project id
		{KeenEventCollection, "", false},       //Keen.io event collection
		{InitStatus, "", false},                //Initialization status used in the init cmd
		{RemoteEnvironmentId, "", false},       //Remote environment Id
		{CpKubeProxyEnabled, "true", false},    //Determine if the Cp Kube proxy is used
		{KubeDirectClusterAddr, "", false},     //Cluster Address (Used only for direct connections to kubernetes)
		{KubeDirectClusterUser, "", false},     //Cluster User (Used only for direct connections to kubernetes)
		{KubeDirectClusterPassword, "", false}, //Cluster Password (Used only for direct connections to kubernetes)
	}
	local.viper = viper.New()
	return local
}
