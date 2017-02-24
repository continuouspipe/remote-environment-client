package mocks

import "github.com/continuouspipe/remote-environment-client/config"

//Mock for Writer
type MockWriter struct {
	write func(p []byte) (n int, err error)
}

func NewMockWriter() *MockWriter {
	return &MockWriter{}
}

func (m *MockWriter) MockWrite(mocked func(p []byte) (n int, err error)) {
	m.write = mocked
}

func (m *MockWriter) Write(p []byte) (n int, err error) {
	return m.write(p)
}

//Mock for Config
type MockConfig struct {
	set            func(key string, value interface{}) error
	getBool        func(key string) (bool, error)
	getString      func(key string) (string, error)
	setConfigFile  func(configType config.ConfigType, in string) error
	setConfigPath  func(configType config.ConfigType, in string) error
	configFileUsed func(configType config.ConfigType) (string, error)
	readInConfig   func(configType config.ConfigType) error
	save           func() error
}

func NewMockConfig() *MockConfig {
	return &MockConfig{}
}

//callers to mocked functions
func (m MockConfig) Set(key string, value interface{}) error {
	return m.set(key, value)
}

func (m MockConfig) GetBool(key string) (bool, error) {
	return m.getBool(key)
}

func (m MockConfig) GetString(key string) (string, error) {
	return m.getString(key)
}

func (m MockConfig) SetConfigFile(configType config.ConfigType, in string) error {

	return m.setConfigFile(configType, in)
}

func (m MockConfig) SetConfigPath(configType config.ConfigType, in string) error {

	return m.setConfigPath(configType, in)
}

func (m MockConfig) ConfigFileUsed(configType config.ConfigType) (string, error) {

	return m.configFileUsed(configType)
}

func (m MockConfig) ReadInConfig(configType config.ConfigType) error {

	return m.readInConfig(configType)
}

func (m MockConfig) Save() error {
	return m.save()
}

//setters for mock functions
func (m *MockConfig) MockSet(mocked func(key string, value interface{}) error) {
	m.set = mocked
}

func (m *MockConfig) MockGetBool(mocked func(key string) (bool, error)) {
	m.getBool = mocked
}
func (m *MockConfig) MockGetString(mocked func(key string) (string, error)) {
	m.getString = mocked
}

func (m *MockConfig) MockSetConfigFile(mocked func(configType config.ConfigType, in string) error) {
	m.setConfigFile = mocked
}

func (m *MockConfig) MockSetConfigPath(mocked func(configType config.ConfigType, in string) error) {
	m.setConfigPath = mocked
}

func (m *MockConfig) MockConfigFileUsed(mocked func(configType config.ConfigType) (string, error)) {
	m.configFileUsed = mocked
}

func (m *MockConfig) MockReadInConfig(mocked func(configType config.ConfigType) error) {
	m.readInConfig = mocked
}

func (m *MockConfig) MockSave(mocked func() error) {
	m.save = mocked
}
