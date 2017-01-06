package config

import (
	"path/filepath"
	"os"
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

	//TODO: write on file
	_ = f
	return true;
}
