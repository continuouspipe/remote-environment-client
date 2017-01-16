package cmd

import (
	"testing"

	"k8s.io/client-go/pkg/api/v1"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/test"
)

func TestSysCallIsCalledToOpenBashSession(t *testing.T) {
	bashHandle := &BashHandle{}

	configReader := test.GetMockConfigReader()
	configReader.MockGetString(func(key string) string {
		switch key {
		case config.KubeConfigKey:
			return "my-config-key"

		case config.Environment:
			return "feature-testing"

		case config.Service:
			return "web"
		}
		return ""
	})

	mockPodsFinder := test.GetMockPodsFinder()
	mockPodsFinder.MockFindAll(func(kubeConfigKey string, environment string) (*v1.PodList, error) {
		return &v1.PodList{}, nil
	})

	mockPodFilter := test.GetMockPodsFilter()
	mockPodFilter.MockByService(func(podList *v1.PodList, service string) (*v1.Pod, error) {
		mockPod := &v1.Pod{}
		mockPod.SetName("web-123456")
		return mockPod, nil
	})

	mockLocalExecutor := test.GetMockLocalExecutor()

	bashHandle.Handle([]string{}, configReader, mockPodsFinder, mockPodFilter, mockLocalExecutor)

	firstCall := mockLocalExecutor.FirstCallsFor("SysCallExec")

	if mockLocalExecutor.CallsCountFor("SysCallExec") != 1 {
		t.Error("Expected SysCallExec to be called only once")
	}

	if str, ok := firstCall.Arguments["kubeConfigKey"].(string); ok {
		test.AssertSame(t, "my-config-key", str)
	} else {
		t.Fatalf("Expected kube config to be a string, given %T", firstCall.Arguments["kubeConfigKey"])
	}

	if str, ok := firstCall.Arguments["environment"].(string); ok {
		test.AssertSame(t, "feature-testing", str)
	} else {
		t.Fatalf("Expected feature testing to be a string, given %T", firstCall.Arguments["environment"])
	}

	if str, ok := firstCall.Arguments["pod"].(string); ok {
		test.AssertSame(t, "web-123456", str)
	} else {
		t.Fatalf("Expected pod to be a string, given %T", firstCall.Arguments["pod"])
	}
}
