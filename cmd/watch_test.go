package cmd

import (
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/continuouspipe/remote-environment-client/sync/monitor"
	"github.com/continuouspipe/remote-environment-client/test"
	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"k8s.io/client-go/pkg/api/v1"
	"os"
	"testing"
)

func TestWatch(t *testing.T) {
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

	//test subject called
	handler := &WatchHandle{}
	handler.kubeConfigKey = "my-config-key"
	handler.ProjectKey = "proj"
	handler.RemoteBranch = "feature-testing"
	handler.Service = "web"
	handler.Latency = 1000
	handler.Stdout = mockStdout
	handler.IndividualFileSyncThreshold = 20
	handler.Handle(spyOsDirectoryMonitor, mockPodsFinder, mockPodFilter)

	//expectations
	firstCall := spyOsDirectoryMonitor.FirstCallsFor("AnyEventCall")

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error: %s", err.Error())
	}

	if spyOsDirectoryMonitor.CallsCountFor("AnyEventCall") != 1 {
		t.Error("Expected AnyEventCall to be called only once")
	}
	if str, ok := firstCall.Arguments["directory"].(string); ok {
		test.AssertSame(t, cwd, str)
	} else {
		t.Fatalf("Expected directory to be a string, given %T", firstCall.Arguments["directory"])
	}

	if observer, ok := firstCall.Arguments["observer"].(*sync.Syncer); ok {
		test.AssertSame(t, observer.Environment, "proj-feature-testing")
		test.AssertSame(t, observer.IndividualFileSyncThreshold, 20)
		test.AssertSame(t, observer.KubeConfigKey, "my-config-key")
		test.AssertSame(t, observer.Pod.GetName(), "web-123456")
	} else {
		t.Fatalf("Expected observer to implement sync.DirectoryEventSyncAll, given %T", firstCall.Arguments["observer"])
	}
}
