package cmd

import (
	"testing"

	"k8s.io/client-go/pkg/api/v1"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/test"
)

func TestSysCallIsCalledToOpenBashSession(t *testing.T) {
	bashHandle := &BashHandle{}

	mockConfigReader := &MockConfigReader{}
	mockPodsFinder := &MockPodsFinder{}
	mockPodFilter := &MockPodsFilter{}
	mockLocalExecutor := &MockLocalExecutor{}

	bashHandle.Handle([]string{}, mockConfigReader, mockPodsFinder, mockPodFilter, mockLocalExecutor)

	test.AssertSame(t, "my-config-key", mockLocalExecutor.kubeConfigKey)
	test.AssertSame(t, "feature-testing", mockLocalExecutor.environment)
	test.AssertSame(t, "web-123456", mockLocalExecutor.pod)
}

type MockConfigReader struct{}

func (m MockConfigReader) GetString(key string) string {
	switch key {
	case config.KubeConfigKey:
		return "my-config-key"

	case config.Environment:
		return "feature-testing"

	case config.Service:
		return "web"
	}
	return ""
}

type MockPodsFinder struct{}

func (m MockPodsFinder) FindAll(kubeConfigKey string, environment string) (*v1.PodList, error) {
	return &v1.PodList{}, nil
}

type MockPodsFilter struct{}

func (m MockPodsFilter) ByService(podList *v1.PodList, service string) (*v1.Pod, error) {
	mockPod := &v1.Pod{}
	mockPod.SetName("web-123456")
	return mockPod, nil
}

type MockLocalExecutor struct {
	callsCount    int
	kubeConfigKey string
	environment   string
	pod           string
	execCmdArgs   []string
}

func (m *MockLocalExecutor) SysCallExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) {
	m.callsCount++
	m.kubeConfigKey = kubeConfigKey
	m.environment = environment
	m.pod = pod
	m.execCmdArgs = execCmdArgs
}
