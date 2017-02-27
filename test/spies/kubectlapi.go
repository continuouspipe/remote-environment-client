package spies

import "k8s.io/client-go/pkg/api/v1"

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
	findAll func(kubeConfigKey string, environment string) (*v1.ServiceList, error)
}

func NewSpyServiceFinder() *SpyServiceFinder {
	return &SpyServiceFinder{}
}

func (s *SpyServiceFinder) FindAll(kubeConfigKey string, environment string) (*v1.ServiceList, error) {
	args := make(Arguments)
	args["kubeConfigKey"] = kubeConfigKey
	args["environment"] = environment

	function := &Function{Name: "FindAll", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.findAll(kubeConfigKey, environment)
}

func (s *SpyServiceFinder) MockFindAll(mocked func(kubeConfigKey string, environment string) (*v1.ServiceList, error)) {
	s.findAll = mocked
}
