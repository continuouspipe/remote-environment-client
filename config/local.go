package config

import (
	"github.com/spf13/viper"
)

type localConfig struct {
	viperWrapper
}

const FlowId = "flow-id"
const ClusterIdentifier = "cluster-identifier"
const KubeEnvironmentName = "kube-environment-name"
const RemoteName = "remote-name"
const RemoteBranch = "remote-branch"
const Service = "service"
const AnybarPort = "anybar-port"
const KeenWriteKey = "keen-write-key"
const KeenProjectId = "keen-project-id"
const KeenEventCollection = "keen-event-collection"
const RemoteEnvironmentId = "remote-environment-id"
const InitStatus = "init-status"

//settings to disable the kube proxy if required
const CpKubeProxyEnabled = "kube-proxy-enabled"
const KubeDirectClusterAddr = "kube-direct-cluster-addr"
const KubeDirectClusterUser = "kube-direct-cluster-user"
const KubeDirectClusterPassword = "kube-direct-cluster-password"

func newLocalConfig() *localConfig {
	local := &localConfig{}
	local.settings = []Setting{
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
