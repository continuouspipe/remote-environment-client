package mocks

import "k8s.io/kubernetes/pkg/api"

//Mock for PodsFilter
type MockPodsFilter struct {
	byService func(*api.PodList, string) (*api.Pod, error)
}

func NewMockPodsFilter() *MockPodsFilter {
	return &MockPodsFilter{}
}

func (m *MockPodsFilter) ByService(podList *api.PodList, service string) (*api.Pod, error) {
	return m.byService(podList, service)
}

func (m *MockPodsFilter) MockByService(mocked func(podList *api.PodList, service string) (*api.Pod, error)) {
	m.byService = mocked
}

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
