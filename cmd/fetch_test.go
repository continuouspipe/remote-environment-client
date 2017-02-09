package cmd

import (
	"testing"

	"fmt"
	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"k8s.io/client-go/pkg/api/v1"
)

func TestFetch(t *testing.T) {
	fmt.Println("Running TestFetch")
	defer fmt.Println("TestFetch Done")

	//get mocked dependencies
	mockPodsFinder := mocks.NewMockPodsFinder()
	mockPodsFinder.MockFindAll(func(kubeConfigKey string, environment string) (*v1.PodList, error) {
		return &v1.PodList{}, nil
	})
	mockPodFilter := mocks.NewMockPodsFilter()
	mockPodFilter.MockByService(func(podList *v1.PodList, service string) (*v1.Pod, error) {
		mockPod := &v1.Pod{}
		mockPod.SetName("web-123456")
		return mockPod, nil
	})
	spyFetcher := spies.NewSpyRsyncFetch()
	spyFetcher.MockFetch(func() error {
		return nil
	})

	//test subject called
	handler := &FetchHandle{}
	handler.kubeConfigKey = "my-config-key"
	handler.ProjectKey = "proj"
	handler.RemoteBranch = "feature-testing"
	handler.Service = "web"
	handler.File = "some-file.txt"
	handler.Handle([]string{}, mockPodsFinder, mockPodFilter, spyFetcher)

	spyFetcher.ExpectsCallCount(t, "Fetch", 1)
	spyFetcher.ExpectsFirstCallArgument(t, "Fetch", "kubeConfigKey", "my-config-key")
	spyFetcher.ExpectsFirstCallArgument(t, "Fetch", "environment", "proj-feature-testing")
	spyFetcher.ExpectsFirstCallArgument(t, "Fetch", "pod", "web-123456")
	spyFetcher.ExpectsFirstCallArgument(t, "Fetch", "filePath", "some-file.txt")
}
