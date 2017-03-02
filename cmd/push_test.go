package cmd

import (
	"testing"

	"fmt"

	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"k8s.io/client-go/pkg/api/v1"
)

func TestPush(t *testing.T) {
	fmt.Println("Running TestPush")
	defer fmt.Println("TestPush Done")

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
	spySyncer := spies.NewSpyRsyncSyncer()
	spySyncer.MockSync(func() error {
		return nil
	})
	spyKubeCtlInitializer := spies.NewSpyKubeCtlInitializer()
	spyKubeCtlInitializer.MockInit(func(environment string) error {
		return nil
	})

	//test subject called
	handler := &PushHandle{}
	handler.kubeCtlInit = spyKubeCtlInitializer
	handler.Environment = "proj-feature-testing"
	handler.Service = "web"
	handler.File = "some-file.txt"
	handler.RemoteProjectPath = "/my/sub/path/"
	handler.Handle([]string{}, mockPodsFinder, mockPodFilter, spySyncer)

	spySyncer.ExpectsCallCount(t, "SetKubeConfigKey", 1)
	spySyncer.ExpectsCallCount(t, "SetRemoteProjectPath", 1)
	spySyncer.ExpectsCallCount(t, "SetEnvironment", 1)
	spySyncer.ExpectsCallCount(t, "SetPod", 1)
	spySyncer.ExpectsCallCount(t, "Sync", 1)
	spySyncer.ExpectsFirstCallArgument(t, "SetKubeConfigKey", "kubeConfigKey", "proj-feature-testing")
	spySyncer.ExpectsFirstCallArgument(t, "SetRemoteProjectPath", "remoteProjectPath", "/my/sub/path/")
	spySyncer.ExpectsFirstCallArgument(t, "SetEnvironment", "environment", "proj-feature-testing")
	spySyncer.ExpectsFirstCallArgument(t, "SetPod", "pod", "web-123456")
	spySyncer.ExpectsFirstCallArgumentStringSlice(t, "Sync", "filePaths", []string{"some-file.txt"})

	spyKubeCtlInitializer.ExpectsCallCount(t, "Init", 1)
}
