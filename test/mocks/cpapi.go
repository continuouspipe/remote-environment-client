package mocks

import (
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/stretchr/testify/mock"
)

//Mock for CpAPIProvider
type MockCpAPIProvider struct {
	mock.Mock
}

func NewMockCpAPIProvider() *MockCpAPIProvider {
	return &MockCpAPIProvider{}
}

func (m *MockCpAPIProvider) SetAPIKey(apiKey string) {
	m.Called(apiKey)
}

func (m *MockCpAPIProvider) GetAPITeams() ([]cpapi.APITeam, error) {
	args := m.Called()
	return args.Get(0).([]cpapi.APITeam), args.Error(1)
}

func (m *MockCpAPIProvider) GetAPIFlows(project string) ([]cpapi.APIFlow, error) {
	args := m.Called()
	return args.Get(0).([]cpapi.APIFlow), args.Error(1)
}

func (m *MockCpAPIProvider) GetAPIUser(user string) (*cpapi.APIUser, error) {
	args := m.Called(user)
	return args.Get(0).(*cpapi.APIUser), args.Error(1)
}

func (m *MockCpAPIProvider) GetAPIEnvironments(flowID string) ([]cpapi.APIEnvironment, error) {
	args := m.Called(flowID)
	return args.Get(0).([]cpapi.APIEnvironment), args.Error(1)
}

func (m *MockCpAPIProvider) GetRemoteEnvironmentStatus(flowID string, environmentID string) (*cpapi.APIRemoteEnvironmentStatus, error) {
	args := m.Called(flowID, environmentID)
	first := args.Get(0).(*cpapi.APIRemoteEnvironmentStatus)
	return first, args.Error(1)
}

func (m *MockCpAPIProvider) RemoteEnvironmentBuild(remoteEnvironmentFlowID string, gitBranch string) error {
	args := m.Called(remoteEnvironmentFlowID, gitBranch)
	return args.Error(0)
}

func (m *MockCpAPIProvider) CancelRunningTide(flowID string, remoteEnvironmentID string) error {
	args := m.Called(flowID, remoteEnvironmentID)
	return args.Error(0)
}

func (m *MockCpAPIProvider) RemoteEnvironmentRunningAndExists(flowID string, environmentID string) (bool, error) {
	args := m.Called(flowID, environmentID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCpAPIProvider) RemoteEnvironmentDestroy(flowID string, environment string, cluster string) error {
	args := m.Called(flowID, environment)
	return args.Error(0)
}

func (m *MockCpAPIProvider) RemoteDevelopmentEnvironmentDestroy(flowID string, remoteEnvironmentID string) error {
	args := m.Called(flowID, remoteEnvironmentID)
	return args.Error(0)
}

func (m *MockCpAPIProvider) CancelTide(tideID string) error {
	args := m.Called(tideID)
	return args.Error(0)
}
