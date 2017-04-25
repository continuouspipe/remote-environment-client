//TODO: Refactor mocks to use testify framework https://github.com/stretchr/testify
package mocks

import (
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/stretchr/testify/mock"
	"k8s.io/kubernetes/pkg/api"
)

//Mock for PodsFilter
type MockPodsFilter struct {
	mock.Mock
}

func NewMockPodsFilter() *MockPodsFilter {
	return &MockPodsFilter{}
}

func (m MockPodsFilter) List(pods api.PodList) pods.Filter {
	m.Called(pods)
	return m
}

func (m MockPodsFilter) ByService(service string) pods.Filter {
	m.Called(service)
	return m
}

func (m MockPodsFilter) ByStatus(status string) pods.Filter {
	m.Called(status)
	return m
}

func (m MockPodsFilter) First() *api.Pod {
	return &api.Pod{}
}

//TODO: Update to use mock.Mock from testify framework https://github.com/stretchr/testify
//Mock for PodsFinder
type MockPodsFinder struct {
	findAll func(user string, apiKey string, address string, environment string) (*api.PodList, error)
}

func NewMockPodsFinder() *MockPodsFinder {
	return &MockPodsFinder{}
}

func (m *MockPodsFinder) FindAll(user string, apiKey string, address string, environment string) (*api.PodList, error) {
	return m.findAll(user, apiKey, address, environment)
}

func (m *MockPodsFinder) MockFindAll(mocked func(user string, apiKey string, address string, environment string) (*api.PodList, error)) {
	m.findAll = mocked
}
