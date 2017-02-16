package config

type GlobalConfig struct {
	settings []Setting
	viperWrapper
}

func NewGlobalConfig() *GlobalConfig {
	global := &GlobalConfig{}
	global.settings = []Setting{
		//CP Username
		{"username", "", true},

		//CP API Key
		{"api-key", "", true}}
	return global
}