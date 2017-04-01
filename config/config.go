package config

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var C *Config

type ConfigType string

const (
	LocalConfigType  ConfigType = "local"
	GlobalConfigType ConfigType = "global"
	AllConfigTypes   ConfigType = "all"
)

type Setting struct {
	Name         string
	DefaultValue string
	Mandatory    bool
}

type ConfigProvider interface {
	Set(key string, value interface{}) error
	GetBool(key string) (bool, error)
	GetString(key string) (string, error)
	SetConfigFile(configType ConfigType, in string) error
	SetConfigPath(configType ConfigType, in string) error
	ConfigFileUsed(configType ConfigType) (string, error)
	ReadInConfig(configType ConfigType) error
	Save(configType ConfigType) error
}

//allows to fetch settings either from global or local config
type Config struct {
	global *globalConfig
	local  *localConfig
}

func NewConfig() *Config {
	c := &Config{}
	c.global = newGlobalConfig()
	c.local = newLocalConfig()
	return c
}

//set the key value on global or local depending who handles it
func (c *Config) Set(key string, value interface{}) error {
	if c.local.HasSetting(key) {
		c.local.Set(key, value)
		return nil
	} else if c.global.HasSetting(key) {
		c.global.Set(key, value)
		return nil
	}
	return fmt.Errorf("The key specified %s didn't match any of the handled configs.", key)
}

//get the bool value on global or local depending who handles it
func (c *Config) GetBool(key string) (bool, error) {
	if c.local.HasSetting(key) {
		return c.local.GetBool(key), nil
	} else if c.global.HasSetting(key) {
		return c.global.GetBool(key), nil
	}
	return false, fmt.Errorf("The key specified %s didn't match any of the handled configs.", key)
}

//get the string value on global or local depending who handles it
func (c *Config) GetString(key string) (string, error) {
	if c.local.HasSetting(key) {
		return c.local.GetString(key), nil
	} else if c.global.HasSetting(key) {
		return c.global.GetString(key), nil
	}
	return "", fmt.Errorf("The key specified %s didn't match any of the handled configs.", key)
}

//set the config file for the given config type
func (c *Config) SetConfigFile(configType ConfigType, in string) error {
	if configType == LocalConfigType {
		c.local.SetConfigFile(in)
		return nil
	} else if configType == GlobalConfigType {
		c.global.SetConfigFile(in)
		return nil
	}
	return fmt.Errorf("The config type specified %s didn't match any of the handled configs.", configType)
}

//set the config path for the given config type
func (c *Config) SetConfigPath(configType ConfigType, in string) error {
	if configType == LocalConfigType {
		c.local.SetConfigPath(in)
		return nil
	} else if configType == GlobalConfigType {
		c.global.SetConfigPath(in)
		return nil
	}
	return fmt.Errorf("The config type specified %s didn't match any of the handled configs.", configType)
}

//set the config file used for the given config type
func (c *Config) ConfigFileUsed(configType ConfigType) (string, error) {
	if configType == LocalConfigType {
		return c.local.ConfigFileUsed(), nil
	} else if configType == GlobalConfigType {
		return c.global.ConfigFileUsed(), nil
	}
	return "", fmt.Errorf("The config type specified %s didn't match any of the handled configs.", configType)
}

//reads from the viper config specified in the configType
func (c *Config) ReadInConfig(configType ConfigType) error {
	if configType == LocalConfigType {
		return c.local.ReadInConfig()
	} else if configType == GlobalConfigType {
		return c.global.ReadInConfig()
	}
	return fmt.Errorf("The config type specified %s didn't match any of the handled configs.", configType)
}

//save the local and global settings on disk
func (c *Config) Save(configType ConfigType) error {
	switch configType {
	case LocalConfigType:
		return c.local.Save()
	case GlobalConfigType:
		return c.global.Save()
	case AllConfigTypes:
		err := c.local.Save()
		if err != nil {
			return err
		}
		return c.global.Save()
	}
	return fmt.Errorf("Specify a config type: %s, %s or %s?", LocalConfigType, GlobalConfigType, AllConfigTypes)
}

//check if all mandatory settings are set for both config types
func (c Config) Validate() (bool, []string) {
	local := c.local.GetMissingMandatorySettings()
	global := c.global.GetMissingMandatorySettings()
	all := append(local, global...)
	return len(all) == 0, all
}

func init() {
	C = NewConfig()
}

//used by local and global config structs to expose the required viper methods
type viperWrapper struct {
	settings   []Setting
	viper      *viper.Viper
	configFile string
}

func (v *viperWrapper) SetConfigFile(in string) {
	v.configFile = in
	v.viper.SetConfigFile(in)
}

func (v *viperWrapper) SetConfigPath(in string) {
	v.viper.AddConfigPath(in)
}

//reads from the viper config, then if the value is missing sets it from the specified default
func (v *viperWrapper) ReadInConfig() error {
	if err := v.viper.ReadInConfig(); err != nil {
		return err
	}
	for _, setting := range v.settings {
		if value := v.viper.GetString(setting.Name); value == "" {
			v.viper.Set(setting.Name, setting.DefaultValue)
		}
	}
	return nil
}

func (v *viperWrapper) Set(key string, value interface{}) {
	v.viper.Set(key, value)
}

func (v viperWrapper) ConfigFileUsed() string {
	return v.viper.ConfigFileUsed()
}

func (v viperWrapper) GetBool(key string) bool {
	return v.viper.GetBool(key)
}

func (v viperWrapper) GetString(key string) string {
	return v.viper.GetString(key)
}

func (v viperWrapper) HasSetting(key string) bool {
	for _, setting := range v.settings {
		if setting.Name == key {
			return true
		}
	}
	return false
}

//saves the settings on disk
func (v viperWrapper) Save() error {
	configFile := v.viper.ConfigFileUsed()

	file, err := os.OpenFile(configFile, os.O_TRUNC|os.O_WRONLY, 0664)
	defer file.Close()
	if err != nil {
		return err
	}

	w := bufio.NewWriter(file)
	for _, setting := range v.settings {
		_, err := w.WriteString(fmt.Sprintf("%s: %s\n", setting.Name, v.GetString(setting.Name)))
		if err != nil {
			return err
		}
	}
	w.Flush()
	return nil
}

func (v viperWrapper) GetMissingMandatorySettings() []string {
	missing := []string{}
	settings := v.GetMandatorySettings()
	for _, setting := range settings {
		if v.GetString(setting) == "" {
			missing = append(missing, setting)
		}
	}
	return missing
}

func (v viperWrapper) GetMandatorySettings() []string {
	settings := []string{}
	for _, setting := range v.settings {
		if setting.Mandatory == true {
			settings = append(settings, setting.Name)
		}
	}
	return settings
}
