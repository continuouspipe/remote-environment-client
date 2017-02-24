package spies

import "github.com/continuouspipe/remote-environment-client/config"

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
	return m.write(p)
}

//Spy for Config
type SpyConfig struct {
	Spy
	set            func(key string, value interface{}) error
	getBool        func(key string) (bool, error)
	getString      func(key string) (string, error)
	setConfigFile  func(configType config.ConfigType, in string) error
	setConfigPath  func(configType config.ConfigType, in string) error
	configFileUsed func(configType config.ConfigType) (string, error)
	readInConfig   func(configType config.ConfigType) error
	save           func() error
}

func NewSpyConfig() *SpyConfig {
	return &SpyConfig{}
}

//callers to mocked functions
func (m *SpyConfig) Set(key string, value interface{}) error {
	args := make(Arguments)
	args["key"] = key
	args["value"] = value

	function := &Function{Name: "Set", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.set(key, value)
}

func (m *SpyConfig) GetBool(key string) (bool, error) {
	args := make(Arguments)
	args["key"] = key

	function := &Function{Name: "GetBool", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.getBool(key)
}

func (m *SpyConfig) GetString(key string) (string, error) {
	args := make(Arguments)
	args["key"] = key

	function := &Function{Name: "GetString", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.getString(key)
}

func (m *SpyConfig) SetConfigFile(configType config.ConfigType, in string) error {
	args := make(Arguments)
	args["configType"] = configType
	args["in"] = in

	function := &Function{Name: "SetConfigFile", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.setConfigFile(configType, in)
}

func (m *SpyConfig) SetConfigPath(configType config.ConfigType, in string) error {
	args := make(Arguments)
	args["configType"] = configType
	args["in"] = in

	function := &Function{Name: "SetConfigPath", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.setConfigPath(configType, in)
}

func (m *SpyConfig) ConfigFileUsed(configType config.ConfigType) (string, error) {
	args := make(Arguments)
	args["configType"] = configType

	function := &Function{Name: "ConfigFileUsed", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.configFileUsed(configType)
}

func (m *SpyConfig) ReadInConfig(configType config.ConfigType) error {
	args := make(Arguments)
	args["configType"] = configType

	function := &Function{Name: "ReadInConfig", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.readInConfig(configType)
}

func (m *SpyConfig) Save() error {
	args := make(Arguments)

	function := &Function{Name: "Save", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.save()
}

//setters for mock functions
func (m *SpyConfig) MockSet(mocked func(key string, value interface{}) error) {
	m.set = mocked
}

func (m *SpyConfig) MockGetBool(mocked func(key string) (bool, error)) {
	m.getBool = mocked
}
func (m *SpyConfig) MockGetString(mocked func(key string) (string, error)) {
	m.getString = mocked
}

func (m *SpyConfig) MockSetConfigFile(mocked func(configType config.ConfigType, in string) error) {
	m.setConfigFile = mocked
}

func (m *SpyConfig) MockSetConfigPath(mocked func(configType config.ConfigType, in string) error) {
	m.setConfigPath = mocked
}

func (m *SpyConfig) MockConfigFileUsed(mocked func(configType config.ConfigType) (string, error)) {
	m.configFileUsed = mocked
}

func (m *SpyConfig) MockReadInConfig(mocked func(configType config.ConfigType) error) {
	m.readInConfig = mocked
}

func (m *SpyConfig) MockSave(mocked func() error) {
	m.save = mocked
}
