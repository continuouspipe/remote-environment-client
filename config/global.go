package config

import (
	"github.com/spf13/viper"
)

type globalConfig struct {
	viperWrapper
}

const (
	//Username on continuous pipe
	Username = "username"

	//ApiKey of the user
	ApiKey = "api-key"

	//CpAuthenticatorApiAddr target address for cp api calls in the format protocol://host:port
	CpAuthenticatorApiAddr = "cp-authenticator-api-addr"

	//CpRiverApiAddr target address for cp api calls in the format protocol://host:port
	CpRiverApiAddr = "cp-river-api-addr"

	//CpKubeProxyAddr target address for cp proxy in the format protocol://host:port
	CpKubeProxyAddr = "cp-kube-proxy-addr"
)

func newGlobalConfig() *globalConfig {
	global := &globalConfig{}
	global.settings = []Setting{
		{Username, "", true},
		{ApiKey, "", true},
		{CpAuthenticatorApiAddr, "https://authenticator.continuouspipe.io", true},
		{CpRiverApiAddr, "https://river.continuouspipe.io", true},
		{CpKubeProxyAddr, "https://kube-proxy.continuouspipe.io", true},
	}
	global.viper = viper.New()
	return global
}
