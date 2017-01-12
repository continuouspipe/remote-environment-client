package config

import (
	"github.com/spf13/viper"
	"html/template"
	"os"
)

const (
	ProjectKey          = "project-key"
	RemoteBranch        = "remote-branch"
	RemoteName          = "remote-name"
	Service             = "service"
	ClusterIp           = "cluster-ip"
	Username            = "username"
	Password            = "password"
	AnybarPort          = "anybar-port"
	KeenWriteKey        = "keen-write-key"
	KeenProjectId       = "keen-project-id"
	KeenEventCollection = "keen-event-collection"
	Environment         = "environment"
	KubeConfigKey       = "kubernetes-config-key"
)

//Contains all remote environment settings
type ApplicationSettings struct {
	//Continuous Pipe project key
	ProjectKey string
	//Name of the git branch used for the remote environment
	RemoteBranch string
	//Github remote name
	RemoteName string
	//default service name for the commands like (watch, bash, fetch and resync)
	DefaultService string
	//IP of the cluster
	ClusterIp string
	//Cluster username
	Username string
	//Cluster password
	Password string
	//Port Number for AnyBar
	AnybarPort string
	//keen.io write key
	KeenWriteKey string
	//keen.io project id
	KeenProjectId string
	//keen.io event collection
	KeenEventCollection string
	//Continuous Pipe project Environment (Called Namespace in Kubernetes)
	Environment string
}

type Writer interface {
	Save(*ApplicationSettings) bool
}

type YamlWriter struct{}

func NewYamlWriter() *YamlWriter {
	return &YamlWriter{}
}

//save on the settings file the config data
func (writer YamlWriter) Save(config *ApplicationSettings) bool {
	absFilePath := viper.ConfigFileUsed()
	if absFilePath == "" {
		return false
	}
	f, err := os.Create(absFilePath)
	if err != nil {
		return false
	}

	tmpl, err := template.New("config").Parse(
		ProjectKey + ": {{.ProjectKey}}\n" +
			RemoteBranch + ": {{.RemoteBranch}}\n" +
			RemoteName + ": {{.RemoteName}}\n" +
			Service + ": {{.DefaultService}}\n" +
			ClusterIp + ": {{.ClusterIp}}\n" +
			Username + ": {{.Username}}\n" +
			Password + ": {{.Password}}\n" +
			AnybarPort + ": {{.AnybarPort}}\n" +
			KeenWriteKey + ": {{.KeenWriteKey}}\n" +
			KeenProjectId + ": {{.KeenProjectId}}\n" +
			KeenEventCollection + ": {{.KeenEventCollection}}\n" +
			Environment + ": {{.Environment}}\n" +
			"# Do not change the kubernetes-config-key. If it has been changed please run the setup command again\n" +
			KubeConfigKey + ": {{.Environment}}\n")
	err = tmpl.Execute(f, config)
	return err == nil
}

//takes the application config from viper and checks that all the required fields are populated
func Validate() (n int, missing []string) {
	var mandatorySettings = []string{"project-key", "remote-branch", "remote-name", "cluster-ip", "service", "username", "password", "kubernetes-config-key"}
	for _, setting := range mandatorySettings {
		if settingValue := viper.Get(setting); settingValue == nil {
			missing = append(missing, setting)
		}
	}
	return len(missing), missing
}
