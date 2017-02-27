package config

import (
	"github.com/spf13/viper"
)

type localConfig struct {
	viperWrapper
}

const Project = "project"
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
const RemoteEnvironmentConfigModifiedAt = "remote-environment-config-modified-at"

func newLocalConfig() *localConfig {
	local := &localConfig{}
	local.settings = []Setting{
		{Project, "", true},              //CP project name (previously Team Name)
		{ClusterIdentifier, "", true},    //CP cluster Identifier
		{KubeEnvironmentName, "", true},  //CP cluster Identifier
		{RemoteName, "origin", true},     //Github remote name (origin by default)
		{RemoteBranch, "", true},         //Git name of the git branch used for the remote environment
		{Service, "web", true},           //Kubernetes service name for the commands like (watch, bash, fetch and resync)
		{AnybarPort, "", false},          //AnyBar port number
		{KeenWriteKey, "", false},        //Keen.io write key
		{KeenProjectId, "", false},       //Keen.io project id
		{KeenEventCollection, "", false}, //Keen.io event collection
		{InitStatus, "", false},          //Initialization status used in the init cmd
		{RemoteEnvironmentId, "", false}, //Remote environment Id

		//Timestamp that indicates when was the last time that the user has modified the remote environment config settings
		//when the timestamp stored locally differs from the one in the server, it means that the config are out of sync and may need re-synced
		{RemoteEnvironmentConfigModifiedAt, "", false},
	}
	local.viper = viper.New()
	return local
}
