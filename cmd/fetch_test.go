//TODO: Refactor test to testify framework https://github.com/stretchr/testify
package cmd

import (
	"testing"

	"fmt"

	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"k8s.io/kubernetes/pkg/api"
)

func TestFetch(t *testing.T) {
	fmt.Println("Running TestFetch")
	defer fmt.Println("TestFetch Done")

	//get mocked dependencies
	mockPodsFinder := mocks.NewMockPodsFinder()
	mockPodsFinder.MockFindAll(func(user string, apiKey string, address string, environment string) (*api.PodList, error) {
		return &api.PodList{}, nil
	})
	mockPodFilter := mocks.NewMockPodsFilter()
	//mockPodFilter.MockByService(func(podList *api.PodList, service string) (*api.Pod, error) {
	//	mockPod := &api.Pod{}
	//	mockPod.SetName("web-123456")
	//	return mockPod, nil
	//})
	spyFetcher := spies.NewSpyRsyncFetch()
	spyKubeCtlInitializer := spies.NewSpyKubeCtlInitializer()
	spyKubeCtlInitializer.MockGetSettings(func() (addr string, user string, apiKey string, err error) {
		return "", "", "", nil
	})

	//test subject called
	handler := &FetchHandle{}
	handler.kubeCtlInit = spyKubeCtlInitializer
	handler.Environment = "proj-feature-testing"
	handler.Service = "web"
	handler.File = "some-file.txt"
	handler.RemoteProjectPath = "/my/sub/path/"
	handler.Handle([]string{}, mockPodsFinder, mockPodFilter, spyFetcher)

	spyFetcher.AssertCalled(t, "SetKubeConfigKey")
	spyFetcher.AssertCalled(t, "SetRemoteProjectPath")
	spyFetcher.AssertCalled(t, "SetEnvironment")
	spyFetcher.AssertCalled(t, "SetPod")
	spyFetcher.AssertCalled(t, "Fetch")
	spyFetcher.AssertCalled(t, "SetKubeConfigKey", "kubeConfigKey", "proj-feature-testing")
	spyFetcher.AssertCalled(t, "SetRemoteProjectPath", "remoteProjectPath", "/my/sub/path/")
	spyFetcher.AssertCalled(t, "SetEnvironment", "environment", "proj-feature-testing")
	spyFetcher.AssertCalled(t, "SetPod", "pod", "web-123456")
	spyFetcher.AssertCalled(t, "Fetch", "filePath", "some-file.txt")

	spyKubeCtlInitializer.ExpectsCallCount(t, "GetSettings", 1)
}
