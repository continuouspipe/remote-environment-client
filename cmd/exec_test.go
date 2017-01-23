package cmd

import (
	"testing"

	"github.com/continuouspipe/remote-environment-client/test"
	"k8s.io/client-go/pkg/api/v1"
)

func TestCommandsAreSpawned(t *testing.T) {
	//get mocked dependencies
	configReader := getMockConfigReaderInitialized()
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
	spyLocalExecutor := test.GetSpyLocalExecutor()
	spyLocalExecutor.SpyCommandExec(func() string {
		return "some retult back.."
	})

	//test subject called
	execHandle := &ExecHandle{}
	out := execHandle.Handle([]string{"ls", "-a", "-l", "-l"}, configReader, mockPodsFinder, mockPodFilter, spyLocalExecutor)

	//expectations
	test.AssertSame(t, "some retult back..", out)
	firstCall := spyLocalExecutor.FirstCallsFor("CommandExec")

	if spyLocalExecutor.CallsCountFor("CommandExec") != 1 {
		t.Error("Expected CommandExec to be called only once")
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
	if execArgs, ok := firstCall.Arguments["execCmdArgs"].([]string); ok {
		test.AssertDeepEqual(t, execArgs, []string{"ls", "-a", "-l", "-l"})
	} else {
		t.Fatalf("Expected pod to be a string, given %T", firstCall.Arguments["pod"])
	}
}
