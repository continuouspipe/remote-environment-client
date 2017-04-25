package mocks

import (
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/errors"
	"github.com/stretchr/testify/mock"
)

//Mock for CpApiProvider
type MockCpApiProvider struct {
	mock.Mock
}

func NewMockCpApiProvider() *MockCpApiProvider {
	return &MockCpApiProvider{}
}

func (m *MockCpApiProvider) SetApiKey(apiKey string) {
	m.Called(apiKey)
}

func (m *MockCpApiProvider) GetApiTeams() ([]cpapi.ApiTeam, error) {
	args := m.Called()
	return args.Get(0).([]cpapi.ApiTeam), args.Error(1)
}

func (m *MockCpApiProvider) GetApiUser(user string) (*cpapi.ApiUser, error) {
	args := m.Called(user)
	return args.Get(0).(*cpapi.ApiUser), args.Error(1)
}

func (m *MockCpApiProvider) GetApiEnvironments(flowId string) ([]cpapi.ApiEnvironment, errors.ErrorListProvider) {
	args := m.Called(flowId)
	return args.Get(0).([]cpapi.ApiEnvironment), args.Get(1).(errors.ErrorListProvider)
}

func (m *MockCpApiProvider) GetRemoteEnvironmentStatus(flowId string, environmentId string) (*cpapi.ApiRemoteEnvironmentStatus, errors.ErrorListProvider) {
	args := m.Called(flowId, environmentId)
	first := args.Get(0).(*cpapi.ApiRemoteEnvironmentStatus)
	if el, ok := args.Get(1).(*errors.ErrorList); ok {
		return first, el
	}
	return first, nil
}

func (m *MockCpApiProvider) RemoteEnvironmentBuild(remoteEnvironmentFlowID string, gitBranch string) error {
	args := m.Called(remoteEnvironmentFlowID, gitBranch)
	return args.Error(0)
}

func (m *MockCpApiProvider) CancelRunningTide(flowId string, remoteEnvironmentId string) error {
	args := m.Called(flowId, remoteEnvironmentId)
	return args.Error(0)
}

func (m *MockCpApiProvider) RemoteEnvironmentRunningAndExists(flowId string, environmentId string) (bool, errors.ErrorListProvider) {
	args := m.Called(flowId, environmentId)
	return args.Bool(0), args.Get(1).(errors.ErrorListProvider)
}

func (m *MockCpApiProvider) RemoteEnvironmentDestroy(flowId string, environment string, cluster string) error {
	args := m.Called(flowId, environment)
	return args.Error(0)
}

func (m *MockCpApiProvider) RemoteDevelopmentEnvironmentDestroy(flowId string, remoteEnvironmentId string) error {
	args := m.Called(flowId, remoteEnvironmentId)
	return args.Error(0)
}

func (m *MockCpApiProvider) CancelTide(tideId string) error {
	args := m.Called(tideId)
	return args.Error(0)
}
