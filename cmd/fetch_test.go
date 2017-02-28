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
	spyKubeCtlInitializer := spies.NewSpyKubeCtlInitializer()
	spyKubeCtlInitializer.MockInit(func() error {
		return nil
	})

	//test subject called
	handler := &FetchHandle{}
	handler.kubeCtlInit = spyKubeCtlInitializer
	handler.Environment = "proj-feature-testing"
	handler.Service = "web"
	handler.File = "some-file.txt"
	handler.RemoteProjectPath = "/my/sub/path/"
	handler.Handle([]string{}, mockPodsFinder, mockPodFilter, spyFetcher)

	spyFetcher.ExpectsCallCount(t, "SetKubeConfigKey", 1)
	spyFetcher.ExpectsCallCount(t, "SetRemoteProjectPath", 1)
	spyFetcher.ExpectsCallCount(t, "SetEnvironment", 1)
	spyFetcher.ExpectsCallCount(t, "SetPod", 1)
	spyFetcher.ExpectsCallCount(t, "Fetch", 1)
	spyFetcher.ExpectsFirstCallArgument(t, "SetKubeConfigKey", "kubeConfigKey", "proj-feature-testing")
	spyFetcher.ExpectsFirstCallArgument(t, "SetRemoteProjectPath", "remoteProjectPath", "/my/sub/path/")
	spyFetcher.ExpectsFirstCallArgument(t, "SetEnvironment", "environment", "proj-feature-testing")
	spyFetcher.ExpectsFirstCallArgument(t, "SetPod", "pod", "web-123456")
	spyFetcher.ExpectsFirstCallArgument(t, "Fetch", "filePath", "some-file.txt")

	spyKubeCtlInitializer.ExpectsCallCount(t, "Init", 1)
}
