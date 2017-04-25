//TODO: Refactor test to testify framework https://github.com/stretchr/testify
package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/sync/monitor"
	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"github.com/stretchr/testify/mock"
	"k8s.io/kubernetes/pkg/api"
	"testing"
	"time"
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
	//mockPodFilter.MockByService(func(podList *api.PodList, service string) (*api.Pod, error) {
	//	mockPod := &api.Pod{}
	//	mockPod.SetName("web-123456")
	//	return mockPod, nil
	//})

	spyOsDirectoryMonitor := spies.NewSpyOsDirectoryMonitor()
	spyOsDirectoryMonitor.MockAnyEventCall(func(directory string, observer monitor.EventsObserver) error {
		return nil
	})

	mockStdout := spies.NewSpyWriter()
	mockStdout.MockWrite(func(p []byte) (n int, err error) {
		return 100, nil
	})

	spySyncer := spies.NewSpySyncer()

	spyKubeCtlInitializer := spies.NewSpyKubeCtlInitializer()
	spyKubeCtlInitializer.MockGetSettings(func() (addr string, user string, apiKey string, err error) {
		return "", "", "", nil
	})
	spyApiProvider := spies.NewSpyApiProvider()
	spyApiProvider.MockGetRemoteEnvironmentStatus(func(flowId string, environmentId string) (*cpapi.ApiRemoteEnvironmentStatus, errors.ErrorListProvider) {
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
	spyConfig.On("GetString", mock.AnythingOfType("string")).Return(func(key string) (string, error) {
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
	handler.options = watchCmdOptions{
		environment:                 "proj-feature-testing",
		remoteProjectPath:           "/my/sub/path/",
		service:                     "web",
		latency:                     1000,
		individualFileSyncThreshold: 20,
	}
	handler.Stdout = mockStdout
	handler.syncer = spySyncer
	handler.config = spyConfig
	handler.api = spyApiProvider
	handler.Handle(spyOsDirectoryMonitor, mockPodsFinder, mockPodFilter)

	//expectations
	spySyncer.AssertNumberOfCalls(t, "SetKubeConfigKey", 1)
	spySyncer.AssertNumberOfCalls(t, "SetEnvironment", 1)
	spySyncer.AssertNumberOfCalls(t, "SetPod", 1)
	spySyncer.AssertNumberOfCalls(t, "SetIndividualFileSyncThreshold", 1)
	spySyncer.AssertNumberOfCalls(t, "SetRemoteProjectPath", 1)

	spySyncer.AssertCalled(t, "SetKubeConfigKey", "key", "proj-feature-testing")
	spySyncer.AssertCalled(t, "SetEnvironment", "env", "proj-feature-testing")
	spySyncer.AssertCalled(t, "SetPod", "pod", "web-123456")
	spySyncer.AssertCalled(t, "SetIndividualFileSyncThreshold", "threshold", 20)
	spySyncer.AssertCalled(t, "SetRemoteProjectPath", "remoteProjectPath", "/my/sub/path/")

	spyOsDirectoryMonitor.ExpectsCallCount(t, "SetLatency", 1)
	spyOsDirectoryMonitor.ExpectsCallCount(t, "AnyEventCall", 1)

	spyOsDirectoryMonitor.ExpectsFirstCallArgument(t, "SetLatency", "latency", time.Duration(1000))

	spyKubeCtlInitializer.ExpectsCallCount(t, "GetSettings", 1)
}
