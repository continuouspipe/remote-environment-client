package test

import "k8s.io/client-go/pkg/api/v1"

//Mock for ConfigReader
type MockConfigReader struct {
	getString func(string) string
}

func GetMockConfigReader() *MockConfigReader {
	return &MockConfigReader{}
}

func (m *MockConfigReader) GetString(key string) string {
	return m.getString(key)
}

func (m *MockConfigReader) MockGetString(mocked func(string) string) {
	m.getString = mocked
}

//Mock for PodsFinder
type MockPodsFinder struct {
	findAll func(kubeConfigKey string, environment string) (*v1.PodList, error)
}

func GetMockPodsFinder() *MockPodsFinder {
	return &MockPodsFinder{}
}

func (m *MockPodsFinder) FindAll(kubeConfigKey string, environment string) (*v1.PodList, error) {
	return m.findAll(kubeConfigKey, environment)
}

func (m *MockPodsFinder) MockFindAll(mocked func(kubeConfigKey string, environment string) (*v1.PodList, error)) {
	m.findAll = mocked
}

//Mock for PodsFilter
type MockPodsFilter struct {
	byService func(podList *v1.PodList, service string) (*v1.Pod, error)
}

func GetMockPodsFilter() *MockPodsFilter {
	return &MockPodsFilter{}
}

func (m *MockPodsFilter) ByService(podList *v1.PodList, service string) (*v1.Pod, error) {
	return m.byService(podList, service)
}

func (m *MockPodsFilter) MockByService(mocked func(podList *v1.PodList, service string) (*v1.Pod, error)) {
	m.byService = mocked
}