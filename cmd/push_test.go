package cmd

import (
	"testing"

	"fmt"

	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"k8s.io/kubernetes/pkg/api"
	"path/filepath"
)

func TestPush(t *testing.T) {
	fmt.Println("Running TestPush")
	defer fmt.Println("TestPush Done")

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
	spySyncer := spies.NewSpyRsyncSyncer()

	spyKubeCtlInitializer := spies.NewSpyKubeCtlInitializer()
	spyKubeCtlInitializer.MockGetSettings(func() (addr string, user string, apiKey string, err error) {
		return "", "", "", nil
	})
	//test subject called
	handler := &PushHandle{}
	handler.kubeCtlInit = spyKubeCtlInitializer

	handler.options = pushCmdOptions{
		environment:       "proj-feature-testing",
		service:           "web",
		file:              "some-file.txt",
		remoteProjectPath: "/my/sub/path/",
	}
	handler.Handle([]string{}, mockPodsFinder, mockPodFilter, spySyncer)

	absFilePath, _ := filepath.Abs("some-file.txt")

	spySyncer.AssertCalled(t, "SetKubeConfigKey")
	spySyncer.AssertCalled(t, "SetRemoteProjectPath")
	spySyncer.AssertCalled(t, "SetEnvironment")
	spySyncer.AssertCalled(t, "SetPod")
	spySyncer.AssertCalled(t, "Sync")
	spySyncer.AssertCalled(t, "SetKubeConfigKey", "kubeConfigKey", "proj-feature-testing")
	spySyncer.AssertCalled(t, "SetRemoteProjectPath", "remoteProjectPath", "/my/sub/path/")
	spySyncer.AssertCalled(t, "SetEnvironment", "environment", "proj-feature-testing")
	spySyncer.AssertCalled(t, "SetPod", "pod", "web-123456")
	spySyncer.AssertCalled(t, "Sync", "filePaths", []string{absFilePath})

	spyKubeCtlInitializer.ExpectsCallCount(t, "GetSettings", 1)
}
