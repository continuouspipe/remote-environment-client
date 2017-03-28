package spies

import (
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/errors"
)

type SpyApiProvider struct {
	Spy
	getApiTeams                         func() ([]cpapi.ApiTeam, error)
	getApiUser                          func(user string) (*cpapi.ApiUser, error)
	getApiEnvironments                  func(flowId string) ([]cpapi.ApiEnvironment, errors.ErrorListProvider)
	getRemoteEnvironmentStatus          func(flowId string, environmentId string) (*cpapi.ApiRemoteEnvironmentStatus, errors.ErrorListProvider)
	remoteEnvironmentBuild              func(remoteEnvironmentFlowID string, gitBranch string) error
	remoteEnvironmentRunningAndExists   func(flowId string, environmentId string) (bool, errors.ErrorListProvider)
	cancelRunningTide                   func(flowId string, remoteEnvironmentId string) error
	remoteEnvironmentDestroy            func(flowId string, environment string, cluster string) error
	remoteDevelopmentEnvironmentDestroy func(flowId string, remoteEnvironmentId string) error
	cancelTide                          func(tideId string) error
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

func (s *SpyApiProvider) GetApiUser(user string) (*cpapi.ApiUser, error) {
	args := make(Arguments)
	args["user"] = user

	function := &Function{Name: "GetApiUser", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.getApiUser(user)
}

func (s *SpyApiProvider) GetApiEnvironments(flowId string) ([]cpapi.ApiEnvironment, errors.ErrorListProvider) {
	args := make(Arguments)
	args["flowId"] = flowId

	function := &Function{Name: "GetApiEnvironments", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.getApiEnvironments(flowId)
}

func (s *SpyApiProvider) GetRemoteEnvironmentStatus(flowId string, environmentId string) (*cpapi.ApiRemoteEnvironmentStatus, errors.ErrorListProvider) {
	args := make(Arguments)
	args["flowId"] = flowId
	args["environmentId"] = environmentId

	function := &Function{Name: "GetRemoteEnvironmentStatus", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.getRemoteEnvironmentStatus(flowId, environmentId)
}

func (s *SpyApiProvider) RemoteEnvironmentBuild(remoteEnvironmentFlowID string, gitBranch string) error {
	args := make(Arguments)
	args["remoteEnvironmentFlowID"] = remoteEnvironmentFlowID
	args["gitBranch"] = gitBranch

	function := &Function{Name: "RemoteEnvironmentBuild", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.remoteEnvironmentBuild(remoteEnvironmentFlowID, gitBranch)
}

func (s *SpyApiProvider) RemoteEnvironmentRunningAndExists(flowId string, environmentId string) (bool, errors.ErrorListProvider) {
	args := make(Arguments)
	args["flowId"] = flowId
	args["environmentId"] = environmentId

	function := &Function{Name: "RemoteEnvironmentRunningAndExists", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)

	if s.remoteEnvironmentRunningAndExists != nil {
		return s.remoteEnvironmentRunningAndExists(flowId, environmentId)
	}
	return false, nil
}

func (s *SpyApiProvider) RemoteDevelopmentEnvironmentDestroy(flowId string, remoteEnvironmentId string) error {
	args := make(Arguments)
	args["flowId"] = flowId
	args["remoteEnvironmentId"] = remoteEnvironmentId

	function := &Function{Name: "RemoteDevelopmentEnvironmentDestroy", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.remoteDevelopmentEnvironmentDestroy(flowId, remoteEnvironmentId)
}

func (s *SpyApiProvider) CancelRunningTide(flowId string, remoteEnvironmentId string) error {
	args := make(Arguments)
	args["flowId"] = flowId
	args["remoteEnvironmentId"] = remoteEnvironmentId

	function := &Function{Name: "CancelRunningTide", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.cancelRunningTide(flowId, remoteEnvironmentId)
}

func (s *SpyApiProvider) RemoteEnvironmentDestroy(flowId string, environment string, cluster string) error {
	args := make(Arguments)
	args["flowId"] = flowId
	args["environment"] = environment
	args["cluster"] = cluster

	function := &Function{Name: "RemoteEnvironmentDestroy", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.remoteEnvironmentDestroy(flowId, environment, cluster)
}

func (s *SpyApiProvider) CancelTide(tideId string) error {
	args := make(Arguments)
	args["tideId"] = tideId

	function := &Function{Name: "CancelTide", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.cancelTide(tideId)
}

func (s *SpyApiProvider) MockGetApiTeams(mocked func() ([]cpapi.ApiTeam, error)) {
	s.getApiTeams = mocked
}

func (s *SpyApiProvider) MockGetApiUser(mocked func(user string) (*cpapi.ApiUser, error)) {
	s.getApiUser = mocked
}

func (s *SpyApiProvider) MockGetApiEnvironments(mocked func(flowId string) ([]cpapi.ApiEnvironment, errors.ErrorListProvider)) {
	s.getApiEnvironments = mocked
}

func (s *SpyApiProvider) MockGetRemoteEnvironmentStatus(mocked func(flowId string, environmentId string) (*cpapi.ApiRemoteEnvironmentStatus, errors.ErrorListProvider)) {
	s.getRemoteEnvironmentStatus = mocked
}

func (s *SpyApiProvider) MockRemoteEnvironmentBuild(mocked func(remoteEnvironmentID string, gitBranch string) error) {
	s.remoteEnvironmentBuild = mocked
}

func (s *SpyApiProvider) MockRemoteEnvironmentRunningAndExists(mocked func(flowId string, environmentId string) (bool, errors.ErrorListProvider)) {
	s.remoteEnvironmentRunningAndExists = mocked
}

func (s *SpyApiProvider) MockRemoteDevelopmentEnvironmentDestroy(mocked func(flowId string, remoteEnvironmentId string) error) {
	s.remoteDevelopmentEnvironmentDestroy = mocked
}

func (s *SpyApiProvider) MockCancelRunningTide(mocked func(flowId string, remoteEnvironmentId string) error) {
	s.cancelRunningTide = mocked
}

func (s *SpyApiProvider) MockRemoteEnvironmentDestroy(mocked func(flowId string, environment string, cluster string) error) {
	s.remoteEnvironmentDestroy = mocked
}

func (s *SpyApiProvider) MockCancelTide(mocked func(tideId string) error) {
	s.cancelTide = mocked
}
