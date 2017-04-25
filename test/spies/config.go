//TODO: Refactor spies to use testify framework https://github.com/stretchr/testify
package spies

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/stretchr/testify/mock"
)

//TODO: Update to use mock.Mock from testify framework https://github.com/stretchr/testify
//Spy for Writer
type SpyWriter struct {
	Spy
	write func(p []byte) (n int, err error)
}

func NewSpyWriter() *SpyWriter {
	return &SpyWriter{}
}

func (m *SpyWriter) MockWrite(mocked func(p []byte) (n int, err error)) {
	m.write = mocked
}

func (m *SpyWriter) Write(p []byte) (n int, err error) {
	args := make(Arguments)
	args["p"] = p

	function := &Function{Name: "Write", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.write(p)
}

//Spy for Config
type SpyConfig struct {
	mock.Mock
}

func NewSpyConfig() *SpyConfig {
	return &SpyConfig{}
}
func (s SpyConfig) Set(key string, value interface{}) error {
	s.Called(key, value)
	return nil
}

func (s SpyConfig) GetBool(key string) (bool, error) {
	s.Called(key)
	return true, nil
}

func (s SpyConfig) GetString(key string) (string, error) {
	s.Called(key)
	return "", nil
}

func (s SpyConfig) SetConfigFile(configType config.ConfigType, in string) error {
	s.Called(configType, in)
	return nil
}

func (s SpyConfig) SetConfigPath(configType config.ConfigType, in string) error {
	s.Called(configType, in)
	return nil
}

func (s SpyConfig) ConfigFileUsed(configType config.ConfigType) (string, error) {
	s.Called(configType)
	return "", nil
}

func (s SpyConfig) ReadInConfig(configType config.ConfigType) error {
	s.Called(configType)
	return nil
}

func (s SpyConfig) Save(configType config.ConfigType) error {
	s.Called(configType)
	return nil
}
