package config

import (
	"github.com/spf13/viper"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

var AppName = os.Args[0]

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

	KubeCtlName = "kubectl"
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

	//Adds configuration reading capabilities
	viper *viper.Viper
}

func NewApplicationSettings() *ApplicationSettings {
	s := &ApplicationSettings{}
	s.ProjectKey = viper.GetString(ProjectKey)
	s.RemoteBranch = viper.GetString(RemoteBranch)
	s.RemoteName = viper.GetString(RemoteName)
	s.DefaultService = viper.GetString(Service)
	s.ClusterIp = viper.GetString(ClusterIp)
	s.Username = viper.GetString(Username)
	s.Password = viper.GetString(Password)
	s.AnybarPort = viper.GetString(AnybarPort)
	s.KeenWriteKey = viper.GetString(KeenWriteKey)
	s.KeenProjectId = viper.GetString(KeenProjectId)
	s.KeenEventCollection = viper.GetString(KeenEventCollection)
	s.Environment = viper.GetString(Environment)
	s.viper = viper.GetViper()
	return s
}

func GetEnvironment(projectKey string, remoteBranch string) string {
	environment := strings.Replace(remoteBranch, string(filepath.Separator), "-", -1)
	environment = strings.Replace(environment, "\\", "-", -1)
	environment = projectKey + "-" + environment
	return environment

}

type Reader interface {
	GetString(key string) string
}

func (s ApplicationSettings) GetString(key string) string {
	return s.viper.GetString(key)
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

type Validator interface {
	Validate(Reader) (n int, missing []string)
}

type MandatoryChecker struct {
	settings []string
}

func NewMandatoryChecker() *MandatoryChecker {
	checker := &MandatoryChecker{}
	checker.settings = []string{ProjectKey,
		RemoteBranch,
		RemoteName,
		ClusterIp,
		Service,
		Username,
		Password,
		KubeConfigKey}
	return checker
}

//takes the application config from a config reader and checks that all the required fields are populated
func (checker *MandatoryChecker) Validate(configReader Reader) (n int, missing []string) {
	for _, setting := range checker.settings {
		if settingValue := configReader.GetString(setting); settingValue == "" {
			missing = append(missing, setting)
		}
	}
	return len(missing), missing
}
