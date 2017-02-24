package spies

import "github.com/continuouspipe/remote-environment-client/cpapi"

type SpyApiProvider struct {
	Spy
	getApiTeams            func() ([]cpapi.ApiTeam, error)
	getApiBucketClusters   func(bucketUuid string) ([]cpapi.ApiCluster, error)
	getApiUser             func(user string) (*cpapi.ApiUser, error)
	getRemoteEnvironment   func(remoteEnvironmentID string) (*cpapi.ApiRemoteEnvironment, error)
	remoteEnvironmentBuild func(remoteEnvironmentID string) error
}

func NewSpyApiProvider() *SpyApiProvider {
	return &SpyApiProvider{}
}

func (s *SpyApiProvider) SetApiKey(apiKey string) {
	args := make(Arguments)
	args["apiKey"] = apiKey

	function := &Function{Name: "SetApiKey", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
}

func (s *SpyApiProvider) GetApiTeams() ([]cpapi.ApiTeam, error) {
	args := make(Arguments)

	function := &Function{Name: "GetApiTeams", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.getApiTeams()
}

func (s *SpyApiProvider) GetApiBucketClusters(bucketUuid string) ([]cpapi.ApiCluster, error) {
	args := make(Arguments)
	args["bucketUuid"] = bucketUuid

	function := &Function{Name: "GetApiBucketClusters", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.getApiBucketClusters(bucketUuid)
}

func (s *SpyApiProvider) GetApiUser(user string) (*cpapi.ApiUser, error) {
	args := make(Arguments)
	args["user"] = user

	function := &Function{Name: "GetApiUser", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.getApiUser(user)
}

func (s *SpyApiProvider) GetRemoteEnvironment(remoteEnvironmentID string) (*cpapi.ApiRemoteEnvironment, error) {
	args := make(Arguments)
	args["remoteEnvironmentID"] = remoteEnvironmentID

	function := &Function{Name: "GetRemoteEnvironment", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.getRemoteEnvironment(remoteEnvironmentID)
}

func (s *SpyApiProvider) RemoteEnvironmentBuild(remoteEnvironmentID string) error {
	args := make(Arguments)
	args["remoteEnvironmentID"] = remoteEnvironmentID

	function := &Function{Name: "RemoteEnvironmentBuild", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.remoteEnvironmentBuild(remoteEnvironmentID)
}

func (s *SpyApiProvider) MockGetApiTeams(mocked func() ([]cpapi.ApiTeam, error)) {
	s.getApiTeams = mocked
}

func (s *SpyApiProvider) MockGetApiBucketClusters(mocked func(bucketUuid string) ([]cpapi.ApiCluster, error)) {
	s.getApiBucketClusters = mocked
}

func (s *SpyApiProvider) MockGetApiUser(mocked func(user string) (*cpapi.ApiUser, error)) {
	s.getApiUser = mocked
}

func (s *SpyApiProvider) MockGetRemoteEnvironment(mocked func(remoteEnvironmentID string) (*cpapi.ApiRemoteEnvironment, error)) {
	s.getRemoteEnvironment = mocked
}

func (s *SpyApiProvider) MockRemoteEnvironmentBuild(mocked func(remoteEnvironmentID string) error) {
	s.remoteEnvironmentBuild = mocked
}
