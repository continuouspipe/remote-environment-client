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

//Mock for QuestionPrompt
type MockQuestionPrompt struct{}

func GetMockQuestionPrompt() *MockQuestionPrompt {
	return &MockQuestionPrompt{}
}

func (qp MockQuestionPrompt) ReadString(q string) string {
	questions := [11]struct {
		question, answer string
	}{
		{"What is your Continuous Pipe project key?", "my-project"},
		{"What is the name of the Git branch you are using for your remote environment?", "feature/MYPROJ-312-initial-commit"},
		{"What is your github remote name? (defaults to: origin)", ""},
		{"What is the default container for the watch, bash, fetch and resync commands?", "web"},
		{"What is the IP of the cluster?", "127.0.0.1"},
		{"What is the cluster username?", "root"},
		{"What is the cluster password?", "2e9fik2s9-fds903"},
		{"If you want to use AnyBar, please provide a port number e.g 1738 ?", "6542"},
		{"What is your keen.io write key? (Optional, only needed if you want to record usage stats)", "sk29dj22d882"},
		{"What is your keen.io project id? (Optional, only needed if you want to record usage stats)", "cc3d902idi01"},
		{"What is your keen.io event collection?  (Optional, only needed if you want to record usage stats)", "event-collection"},
	}
	for _, v := range questions {
		if q == v.question {
			return v.answer
		}
	}
	return ""
}

func (qp MockQuestionPrompt) ApplyDefault(question string, predef string) string {
	return predef
}

func (qp MockQuestionPrompt) RepeatIfEmpty(question string) string {
	return qp.ReadString(question)
}
