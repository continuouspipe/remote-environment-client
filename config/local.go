package config

import (
	"strings"
	"path/filepath"
	"github.com/spf13/viper"
)

type localConfig struct {
	viperWrapper
}

const Project = "project"
const Flow = "flow"
const ClusterIdentifier = "cluster-identifier"
const ProjectKey = "project-key"
const RemoteName = "remote-name"
const RemoteBranch = "remote-branch"
const Service = "service"
const KubeConfigKey = "kubernetes-config-key"
const AnybarPort = "anybar-port"
const KeenWriteKey = "keen-write-key"
const KeenProjectId = "keen-project-id"
const KeenEventCollection = "keen-event-collection"

func newLocalConfig() *localConfig {
	local := &localConfig{}
	local.settings = []Setting{
		{Project, "", true},              //CP project name (previously Team Name)
		{Flow, "", true},                 //CP flow Name
		{ClusterIdentifier, "", true},    //CP cluster Identifier
		{ProjectKey, "", true},           //CP project key
		{RemoteName, "origin", true},     //Github remote name (origin by default)
		{RemoteBranch, "", true},         //Git name of the git branch used for the remote environment
		{Service, "web", true},           //Kubernetes service name for the commands like (watch, bash, fetch and resync)
		{KubeConfigKey, "", true},        //Kubernetes configuration file key
		{AnybarPort, "", false},          //AnyBar port number
		{KeenWriteKey, "", false},        //Keen.io write key
		{KeenProjectId, "", false},       //Keen.io project id
		{KeenEventCollection, "", false}, //Keen.io event collection
	}
	local.viper = viper.New()
	for _, setting := range local.settings {
		local.viper.Set(setting.Name, setting.DefaultValue)
	}
	return local
}

func GetEnvironment(projectKey string, remoteBranch string) string {
	environment := strings.Replace(remoteBranch, string(filepath.Separator), "-", -1)
	environment = strings.Replace(environment, "\\", "-", -1)
	environment = projectKey + "-" + environment
	return environment
}
