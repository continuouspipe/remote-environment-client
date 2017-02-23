package config

import (
	"github.com/spf13/viper"
)

type globalConfig struct {
	viperWrapper
}

//Username on continuous pipe
const Username = "username"

//ApiKey of the user
const ApiKey = "api-key"

//CpApiAddr target address for cp api calls in the format host:port
const CpApiAddr = "cp-api-addr"

func newGlobalConfig() *globalConfig {
	global := &globalConfig{}
	global.settings = []Setting{
		{Username, "", true},
		{ApiKey, "", true},
		{CpApiAddr, "https://authenticator.continuouspipe.io", true},
	}
	global.viper = viper.New()
	for _, setting := range global.settings {
		global.viper.Set(setting.Name, setting.DefaultValue)
	}
	return global
}
