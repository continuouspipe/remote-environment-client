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

func (m *MockPodsFilter) List(pods api.PodList) pods.Filter {
	args := m.Called(pods)
	return args.Get(0).(*MockPodsFilter)
}

func (m *MockPodsFilter) ByService(service string) pods.Filter {
	args := m.Called(service)
	return args.Get(0).(*MockPodsFilter)
}

func (m *MockPodsFilter) ByStatus(status string) pods.Filter {
	args := m.Called(status)
	return args.Get(0).(*MockPodsFilter)
}

func (m *MockPodsFilter) First() *api.Pod {
	args := m.Called()
	return args.Get(0).(*api.Pod)
}
