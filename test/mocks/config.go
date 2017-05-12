package mocks

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/stretchr/testify/mock"
)

//Spy for Config
type SpyConfig struct {
	mock.Mock
}

func NewSpyConfig() *SpyConfig {
	return &SpyConfig{}
}
func (s *SpyConfig) Set(key string, value interface{}) error {
	args := s.Called(key, value)
	return args.Error(0)
}

func (s *SpyConfig) GetBool(key string) (bool, error) {
	args := s.Called(key)
	return args.Bool(0), args.Error(1)
}

func (s *SpyConfig) GetString(key string) (string, error) {
	args := s.Called(key)
	return args.String(0), args.Error(1)
}

func (s *SpyConfig) GetStringQ(key string) string {
	args := s.Called(key)
	return args.String(0)
}

func (s *SpyConfig) SetConfigFile(configType config.ConfigType, in string) error {
	args := s.Called(configType, in)
	return args.Error(0)
}

func (s *SpyConfig) SetConfigPath(configType config.ConfigType, in string) error {
	args := s.Called(configType, in)
	return args.Error(0)
}

func (s *SpyConfig) ConfigFileUsed(configType config.ConfigType) (string, error) {
	args := s.Called(configType)
	return args.String(0), args.Error(1)
}

func (s *SpyConfig) ReadInConfig(configType config.ConfigType) error {
	args := s.Called(configType)
	return args.Error(0)
}

func (s *SpyConfig) Save(configType config.ConfigType) error {
	args := s.Called(configType)
	return args.Error(0)
}
