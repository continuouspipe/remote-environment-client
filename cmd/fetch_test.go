package cmd

import (
	"testing"

	"github.com/continuouspipe/remote-environment-client/test"
	"k8s.io/client-go/pkg/api/v1"
)

func TestFetch(t *testing.T) {
	//get mocked dependencies
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
	spyFetcher := test.GetSpyRsyncFetch()

	//test subject called
	handler := &FetchHandle{}
	handler.kubeConfigKey = "my-config-key"
	handler.ProjectKey = "proj"
	handler.RemoteBranch = "feature-testing"
	handler.Service = "web"
	handler.Handle([]string{}, mockPodsFinder, mockPodFilter, spyFetcher)

	//expectations
	firstCall := spyFetcher.FirstCallsFor("Fetch")

	if spyFetcher.CallsCountFor("Fetch") != 1 {
		t.Error("Expected Fetch to be called only once")
	}
	if str, ok := firstCall.Arguments["kubeConfigKey"].(string); ok {
		test.AssertSame(t, "my-config-key", str)
	} else {
		t.Fatalf("Expected kube config to be a string, given %T", firstCall.Arguments["kubeConfigKey"])
	}
	if str, ok := firstCall.Arguments["environment"].(string); ok {
		test.AssertSame(t, "proj-feature-testing", str)
	} else {
		t.Fatalf("Expected feature testing to be a string, given %T", firstCall.Arguments["environment"])
	}
	if str, ok := firstCall.Arguments["pod"].(string); ok {
		test.AssertSame(t, "web-123456", str)
	} else {
		t.Fatalf("Expected pod to be a string, given %T", firstCall.Arguments["pod"])
	}
}
