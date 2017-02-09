package mocks

import "k8s.io/client-go/pkg/api/v1"

//Mock for PodsFilter
type MockPodsFilter struct {
	byService func(podList *v1.PodList, service string) (*v1.Pod, error)
}

func NewMockPodsFilter() *MockPodsFilter {
	return &MockPodsFilter{}
}

func (m *MockPodsFilter) ByService(podList *v1.PodList, service string) (*v1.Pod, error) {
	return m.byService(podList, service)
}

func (m *MockPodsFilter) MockByService(mocked func(podList *v1.PodList, service string) (*v1.Pod, error)) {
	m.byService = mocked
}

//Mock for PodsFinder
type MockPodsFinder struct {
	findAll func(kubeConfigKey string, environment string) (*v1.PodList, error)
}

func NewMockPodsFinder() *MockPodsFinder {
	return &MockPodsFinder{}
}

func (m *MockPodsFinder) FindAll(kubeConfigKey string, environment string) (*v1.PodList, error) {
	return m.findAll(kubeConfigKey, environment)
}

func (m *MockPodsFinder) MockFindAll(mocked func(kubeConfigKey string, environment string) (*v1.PodList, error)) {
	m.findAll = mocked
}
