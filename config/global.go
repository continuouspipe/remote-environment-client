package config

import (
	"github.com/spf13/viper"
)

type globalConfig struct {
	viperWrapper
}

const Username = "username"
const ApiKey = "api-key"

func newGlobalConfig() *globalConfig {
	global := &globalConfig{}
	global.settings = []Setting{
		//CP Username
		{Username, "", true},

		//CP API Key
		{ApiKey, "", true}}
	global.viper = viper.New()
	for _, setting := range global.settings {
		global.viper.Set(setting.Name, setting.DefaultValue)
	}
	return global
}
