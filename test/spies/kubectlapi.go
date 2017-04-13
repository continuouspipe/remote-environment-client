package spies

import "k8s.io/kubernetes/pkg/api"

type SpyKubeCtlConfigProvider struct {
	Spy
	configSetAuthInfo func(environment string, username string, password string) (string, error)
	configSetCluster  func(environment string, clusterIp string, teamName string, clusterIdentifier string) (string, error)
	configSetContext  func(environment string, username string) (string, error)
}

func NewSpyKubeCtlConfigProvider() *SpyKubeCtlConfigProvider {
	return &SpyKubeCtlConfigProvider{}
}

func (s *SpyKubeCtlConfigProvider) ConfigSetAuthInfo(environment string, username string, password string) (string, error) {
	args := make(Arguments)
	args["environment"] = environment
	args["username"] = username
	args["password"] = password

	function := &Function{Name: "ConfigSetAuthInfo", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.configSetAuthInfo(environment, username, password)
}

func (s *SpyKubeCtlConfigProvider) ConfigSetCluster(environment string, clusterIp string, teamName string, clusterIdentifier string) (string, error) {
	args := make(Arguments)
	args["environment"] = environment
	args["clusterIp"] = clusterIp
	args["teamName"] = teamName
	args["clusterIdentifier"] = clusterIdentifier

	function := &Function{Name: "ConfigSetCluster", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.configSetCluster(environment, clusterIp, teamName, clusterIdentifier)
}

func (s *SpyKubeCtlConfigProvider) ConfigSetContext(environment string, username string) (string, error) {
	args := make(Arguments)
	args["environment"] = environment
	args["username"] = username

	function := &Function{Name: "ConfigSetContext", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.configSetContext(environment, username)
}

func (s *SpyKubeCtlConfigProvider) MockConfigSetAuthInfo(mocked func(environment string, username string, password string) (string, error)) {
	s.configSetAuthInfo = mocked
}

func (s *SpyKubeCtlConfigProvider) MockConfigSetCluster(mocked func(environment string, clusterIp string, teamName string, clusterIdentifier string) (string, error)) {
	s.configSetCluster = mocked
}

func (s *SpyKubeCtlConfigProvider) MockConfigSetContext(mocked func(environment string, username string) (string, error)) {
	s.configSetContext = mocked
}

type SpyKubeCtlClusterInfoProvider struct {
	Spy
	clusterInfo func(kubeConfigKey string) (string, error)
}

func NewSpyKubeCtlClusterInfoProvider() *SpyKubeCtlClusterInfoProvider {
	return &SpyKubeCtlClusterInfoProvider{}
}

func (s *SpyKubeCtlClusterInfoProvider) ClusterInfo(kubeConfigKey string) (string, error) {
	args := make(Arguments)
	args["kubeConfigKey"] = kubeConfigKey

	function := &Function{Name: "ClusterInfo", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.clusterInfo(kubeConfigKey)
}

func (s *SpyKubeCtlClusterInfoProvider) MockClusterInfo(mocked func(kubeConfigKey string) (string, error)) {
	s.clusterInfo = mocked
}

type SpyServiceFinder struct {
	Spy
	findAll func(user string, apiKey string, address string, environment string) (*api.ServiceList, error)
}

func NewSpyServiceFinder() *SpyServiceFinder {
	return &SpyServiceFinder{}
}

func (s *SpyServiceFinder) FindAll(user string, apiKey string, address string, environment string) (*api.ServiceList, error) {
	args := make(Arguments)
	args["user"] = user
	args["apiKey"] = apiKey
	args["address"] = address
	args["environment"] = environment

	function := &Function{Name: "FindAll", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.findAll(user, apiKey, address, environment)
}

func (s *SpyServiceFinder) MockFindAll(mocked func(user string, apiKey string, address string, environment string) (*api.ServiceList, error)) {
	s.findAll = mocked
}

type SpyKubeCtlInitializer struct {
	Spy
	init        func(environment string) error
	getSettings func() (addr string, user string, apiKey string, err error)
}

func NewSpyKubeCtlInitializer() *SpyKubeCtlInitializer {
	return &SpyKubeCtlInitializer{}
}

func (s *SpyKubeCtlInitializer) Init(environment string) error {
	args := make(Arguments)
	args["environment"] = environment

	function := &Function{Name: "Init", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.init(environment)
}

func (s *SpyKubeCtlInitializer) GetSettings() (addr string, user string, apiKey string, err error) {
	args := make(Arguments)

	function := &Function{Name: "GetSettings", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.getSettings()
}

func (s *SpyKubeCtlInitializer) MockInit(mocked func(environment string) error) {
	s.init = mocked
}

func (s *SpyKubeCtlInitializer) MockGetSettings(mocked func() (addr string, user string, apiKey string, err error)) {
	s.getSettings = mocked
}