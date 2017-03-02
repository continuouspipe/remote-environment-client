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
	spyPusher := spies.NewSpyRsyncPush()
	spyPusher.MockPush(func() error {
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
	handler.Handle([]string{}, mockPodsFinder, mockPodFilter, spyPusher)

	spyPusher.ExpectsCallCount(t, "SetKubeConfigKey", 1)
	spyPusher.ExpectsCallCount(t, "SetRemoteProjectPath", 1)
	spyPusher.ExpectsCallCount(t, "SetEnvironment", 1)
	spyPusher.ExpectsCallCount(t, "SetPod", 1)
	spyPusher.ExpectsCallCount(t, "Push", 1)
	spyPusher.ExpectsFirstCallArgument(t, "SetKubeConfigKey", "kubeConfigKey", "proj-feature-testing")
	spyPusher.ExpectsFirstCallArgument(t, "SetRemoteProjectPath", "remoteProjectPath", "/my/sub/path/")
	spyPusher.ExpectsFirstCallArgument(t, "SetEnvironment", "environment", "proj-feature-testing")
	spyPusher.ExpectsFirstCallArgument(t, "SetPod", "pod", "web-123456")
	spyPusher.ExpectsFirstCallArgument(t, "Push", "filePath", "some-file.txt")

	spyKubeCtlInitializer.ExpectsCallCount(t, "Init", 1)
}
