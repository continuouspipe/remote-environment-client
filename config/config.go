package config

import "github.com/spf13/viper"

var C *Config

type Setting struct {
	Name      string
	Value     string
	Mandatory bool
}

//allows to fetch settings either from global or local config
type Config struct {
	Global *GlobalConfig
	Local  *LocalConfig
}

func NewConfig() *Config {
	config := &Config{}
	config.Global = NewGlobalConfig()
	config.Local = NewLocalConfig()
	return config
}

func (c *Config) Set(key string, value interface{}) {
	//set the key value on global or local depending who handles it
}

func (c *Config) GetBool(key string) bool {
	//get the bool value on global or local depending who handles it
	return false
}

func (c *Config) GetString(key string) string {
	//get the bool value on global or local depending who handles it
	return ""
}

func init() {
	C = NewConfig()
}

//used by local and global config structs to expose the required viper methods
type viperWrapper struct {
	viper *viper.Viper
}

func (w *viperWrapper) SetConfigFile(in string) {
	w.viper.SetConfigFile(in)
}

func (w *viperWrapper) AddConfigPath(in string) {
	w.viper.AddConfigPath(in)
}

func (w *viperWrapper) ConfigFileUsed() string {
	return w.viper.ConfigFileUsed()
}

func (w *viperWrapper) ReadInConfig() error {
	return w.viper.ReadInConfig()
}

func (w *viperWrapper) Set(key string, value interface{}) {
	w.viper.Set(key, value)
}

func (w *viperWrapper) GetBool(key string) bool {
	return w.viper.GetBool(key)
}

func (w *viperWrapper) GetString(key string) string {
	return w.viper.GetString(key)
}
