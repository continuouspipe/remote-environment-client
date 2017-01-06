package config

import (
	"path/filepath"
	"os"
	"fmt"
	"html/template"
)

const settingsFile = ".cp-remote-env-settings"

//Contains all remote environment settings
type ConfigData struct {
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

//return the fullpath to the settings file
func SettingsFileDir() string {
	p, err := filepath.Abs(settingsFile)
	if err != nil {
		return ""
	}
	return filepath.Clean(p)
}

//save on the settings file the config data
func (config ConfigData) SaveOnDisk() bool {
	absFilePath := SettingsFileDir()
	if absFilePath == "" {
		return false
	}
	f, err := os.Create(absFilePath)
	if err != nil {
		return false;
	}

	tmpl, err := template.New("config").Parse(`"PROJECT_KEY={{.ProjectKey}}"
"REMOTE_BRANCH={{.RemoteBranch}}"
"REMOTE_NAME={{.RemoteName}}"
"DEFAULT_CONTAINER={{.DefaultContainer}}"
"ANYBAR_PORT={{.AnybarPort}}"
"KEEN_WRITE_KEY={{.KeenWriteKey}}"
"KEEN_PROJECT_ID={{.KeenProjectId}}"
"KEEN_EVENT_COLLECTION={{.KeenEventCollection}}"
"# Do not change the KUBERNETES_CONFIG_KEY. If it has been changed please run the setup command again"
"KUBERNETES_CONFIG_KEY={{.Namespace}}"`)

	//TODO: Write template to file
	_ = f
	_ = tmpl
	return true;
}
