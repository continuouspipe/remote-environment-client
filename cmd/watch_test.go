package cmd

import (
	"fmt"
	"testing"
	"time"

	"github.com/continuouspipe/remote-environment-client/sync/monitor"
	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"k8s.io/client-go/pkg/api/v1"
)

func TestWatch(t *testing.T) {
	fmt.Println("Running TestWatch")
	defer fmt.Println("TestWatch Done")

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

	spyOsDirectoryMonitor := spies.NewSpyOsDirectoryMonitor()
	spyOsDirectoryMonitor.MockAnyEventCall(func(directory string, observer monitor.EventsObserver) error {
		return nil
	})

	mockStdout := mocks.NewMockWriter()
	mockStdout.MockWrite(func(p []byte) (n int, err error) {
		return 100, nil
	})

	spySyncer := spies.NewSpySyncer()
	spySyncer.MockSync(func(filePaths []string) error {
		return nil
	})

	//test subject called
	handler := &WatchHandle{}
	handler.Environment = "proj-feature-testing"
	handler.RemoteProjectPath = "/my/sub/path/"
	handler.Service = "web"
	handler.Latency = 1000
	handler.Stdout = mockStdout
	handler.IndividualFileSyncThreshold = 20
	handler.syncer = spySyncer
	handler.Handle(spyOsDirectoryMonitor, mockPodsFinder, mockPodFilter)

	//expectations
	spySyncer.ExpectsCallCount(t, "SetKubeConfigKey", 1)
	spySyncer.ExpectsCallCount(t, "SetEnvironment", 1)
	spySyncer.ExpectsCallCount(t, "SetPod", 1)
	spySyncer.ExpectsCallCount(t, "SetIndividualFileSyncThreshold", 1)
	spySyncer.ExpectsCallCount(t, "SetRemoteProjectPath", 1)

	spySyncer.ExpectsFirstCallArgument(t, "SetKubeConfigKey", "key", "proj-feature-testing")
	spySyncer.ExpectsFirstCallArgument(t, "SetEnvironment", "env", "proj-feature-testing")
	spySyncer.ExpectsFirstCallArgument(t, "SetPod", "pod", "web-123456")
	spySyncer.ExpectsFirstCallArgument(t, "SetIndividualFileSyncThreshold", "threshold", 20)
	spySyncer.ExpectsFirstCallArgument(t, "SetRemoteProjectPath", "remoteProjectPath", "/my/sub/path/")

	spyOsDirectoryMonitor.ExpectsCallCount(t, "SetLatency", 1)
	spyOsDirectoryMonitor.ExpectsCallCount(t, "AnyEventCall", 1)

	spyOsDirectoryMonitor.ExpectsFirstCallArgument(t, "SetLatency", "latency", time.Duration(1000))
}
