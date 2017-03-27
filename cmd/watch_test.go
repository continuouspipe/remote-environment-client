package cmd

import (
	"fmt"
	"testing"
	"time"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/sync/monitor"
	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"k8s.io/kubernetes/pkg/api"
)

func TestWatch(t *testing.T) {
	fmt.Println("Running TestWatch")
	defer fmt.Println("TestWatch Done")

	//get mocked dependencies
	mockPodsFinder := mocks.NewMockPodsFinder()
	mockPodsFinder.MockFindAll(func(user string, apiKey string, address string, environment string) (*api.PodList, error) {
		return &api.PodList{}, nil
	})
	mockPodFilter := mocks.NewMockPodsFilter()
	mockPodFilter.MockByService(func(podList *api.PodList, service string) (*api.Pod, error) {
		mockPod := &api.Pod{}
		mockPod.SetName("web-123456")
		return mockPod, nil
	})

	spyOsDirectoryMonitor := spies.NewSpyOsDirectoryMonitor()
	spyOsDirectoryMonitor.MockAnyEventCall(func(directory string, observer monitor.EventsObserver) error {
		return nil
	})

	mockStdout := spies.NewSpyWriter()
	mockStdout.MockWrite(func(p []byte) (n int, err error) {
		return 100, nil
	})

	spySyncer := spies.NewSpySyncer()
	spySyncer.MockSync(func(filePaths []string) error {
		return nil
	})

	spyKubeCtlInitializer := spies.NewSpyKubeCtlInitializer()
	spyKubeCtlInitializer.MockGetSettings(func() (addr string, user string, apiKey string, err error) {
		return "", "", "", nil
	})
	spyApiProvider := spies.NewSpyApiProvider()
	spyApiProvider.MockGetRemoteEnvironmentStatus(func(flowId string, environmentId string) (*cpapi.ApiRemoteEnvironmentStatus, error) {
		r := &cpapi.ApiRemoteEnvironmentStatus{}
		r.PublicEndpoints = []cpapi.ApiPublicEndpoint{
			{
				Address: "10.0.0.0",
				Name:    "web",
				Ports: []cpapi.ApiPublicEndpointPort{
					{
						Number:   80,
						Protocol: "tcp",
					},
				},
			},
		}
		return r, nil
	})
	spyConfig := spies.NewSpyConfig()
	spyConfig.MockGetString(func(key string) (string, error) {
		switch key {
		case config.ApiKey:
			return "some-api-key", nil
		case config.RemoteEnvironmentId:
			return "987654321", nil
		case config.FlowId:
			return "837d92hd-19su1d91", nil
		}
		return "", nil
	})

	//test subject called
	handler := &WatchHandle{}
	handler.kubeCtlInit = spyKubeCtlInitializer
	handler.Environment = "proj-feature-testing"
	handler.RemoteProjectPath = "/my/sub/path/"
	handler.Service = "web"
	handler.Latency = 1000
	handler.Stdout = mockStdout
	handler.IndividualFileSyncThreshold = 20
	handler.syncer = spySyncer
	handler.config = spyConfig
	handler.api = spyApiProvider
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

	spyKubeCtlInitializer.ExpectsCallCount(t, "GetSettings", 1)
}
