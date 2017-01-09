package config

import (
	"github.com/spf13/viper"
	"html/template"
	"os"
)

//Contains all remote environment settings
type ApplicationSettings struct {
	//Continuous Pipe project key
	ProjectKey string
	//Name of the git branch used for the remote environment
	RemoteBranch string
	//Github remote name
	RemoteName string
	//default container for the commands like (watch, bash, fetch and resync)
	DefaultContainer string
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
	//Continuous Pipe project Namespace
	Namespace string
}

type Writer interface {
	Save(*ApplicationSettings) bool
}

type YamlWriter struct{}

func NewYamlWriter() *YamlWriter {
	return &YamlWriter{};
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

	tmpl, err := template.New("config").Parse(`project-key: {{.ProjectKey}}
remote-branch: {{.RemoteBranch}}
remote-name: {{.RemoteName}}
container: {{.DefaultContainer}}
environment: {{.Namespace}}
anybar-port: {{.AnybarPort}}
keen-write-key: {{.KeenWriteKey}}
keen-project-id: {{.KeenProjectId}}
keen-event-collection: {{.KeenEventCollection}}
# Do not change the kubernetes-config-key. If it has been changed please run the setup command again
kubernetes-config-key: {{.Namespace}}`)

	err = tmpl.Execute(f, config)
	return err == nil
}
