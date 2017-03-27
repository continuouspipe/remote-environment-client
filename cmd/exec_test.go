package cmd

import (
	"testing"

	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/errors"
	kexec "github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"k8s.io/kubernetes/pkg/api"
	"os"
)

func TestCommandsAreSpawned(t *testing.T) {
	fmt.Println("Running TestCommandsAreSpawned")
	defer fmt.Println("TestCommandsAreSpawned Done")

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
	spyLocalExecutor := spies.NewSpyLocalExecutor()
	spyLocalExecutor.MockStartProcess(func() error {
		return nil
	})
	spyKubeCtlInitializer := spies.NewSpyKubeCtlInitializer()
	spyKubeCtlInitializer.MockInit(func(environment string) error {
		return nil
	})
	spyKubeCtlInitializer.MockGetSettings(func() (addr string, user string, apiKey string, err error) {
		return "", "", "", nil
	})
	spyConfig := spies.NewSpyConfig()

	//test subject called
	handler := &execHandle{}
	handler.kubeCtlInit = spyKubeCtlInitializer
	handler.environment = "proj-feature-testing"
	handler.service = "web"
	handler.Complete([]string{"ls", "-a", "-l", "-l"}, spyConfig)
	handler.Handle(mockPodsFinder, mockPodFilter, spyLocalExecutor)

	//expectations
	kscmd := kexec.KSCommand{}
	kscmd.KubeConfigKey = "proj-feature-testing"
	kscmd.Environment = "proj-feature-testing"
	kscmd.Pod = "web-123456"
	kscmd.Stdin = os.Stdin
	kscmd.Stdout = os.Stdout
	kscmd.Stderr = os.Stderr

	spyLocalExecutor.ExpectsCallCount(t, "StartProcess", 1)
	spyLocalExecutor.ExpectsFirstCallArgument(t, "StartProcess", "kscmd", kscmd)
	spyLocalExecutor.ExpectsFirstCallArgumentStringSlice(t, "StartProcess", "execCmdArgs", []string{"ls", "-a", "-l", "-l"})

	spyKubeCtlInitializer.ExpectsCallCount(t, "Init", 1)
}

func TestExecHandle_Handle_InteractiveMode(t *testing.T) {
	fmt.Println("Running TestExecHandle_Handle_InteractiveMode")
	defer fmt.Println("TestExecHandle_Handle_InteractiveMode Done")

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
	spyLocalExecutor := spies.NewSpyLocalExecutor()
	spyLocalExecutor.MockStartProcess(func() error {
		return nil
	})
	spyKubeCtlInitializer := spies.NewSpyKubeCtlInitializer()
	spyKubeCtlInitializer.MockInit(func(environment string) error {
		return nil
	})
	spyKubeCtlInitializer.MockGetSettings(func() (addr string, user string, apiKey string, err error) {
		return "a", "b", "c", nil
	})
	spyConfig := spies.NewSpyConfig()
	spyConfig.MockGetString(func(key string) (string, error) {
		return "", nil
	})
	spyConfig.MockSet(func(key string, value interface{}) error {
		return nil
	})

	spyInitInteractiveH := spies.NewSpyInitStrategy()
	spyInitInteractiveH.MockComplete(func(argsIn []string) error {
		return nil
	})
	spyInitInteractiveH.MockValidate(func() error {
		return nil
	})
	spyInitInteractiveH.MockHandle(func() error {
		return nil
	})

	spyApi := spies.NewSpyApiProvider()
	spyApi.MockGetApiEnvironments(func(flowId string) ([]cpapi.ApiEnvironment, *errors.ErrorList) {
		return []cpapi.ApiEnvironment{
			{
				"my-cluster",
				nil,
				"proj-feature-testing",
			},
		}, nil
	})
	//test subject called
	handler := &execHandle{}
	handler.kubeCtlInit = spyKubeCtlInitializer
	handler.environment = "proj-feature-testing"
	handler.service = "web"
	handler.initInteractiveH = spyInitInteractiveH
	handler.interactive = true
	handler.api = spyApi
	handler.flowId = "my-flow"
	handler.config = spyConfig
	handler.Complete([]string{"ls", "-a", "-l", "-l"}, spyConfig)
	handler.Handle(mockPodsFinder, mockPodFilter, spyLocalExecutor)

	//expectations
	kscmd := kexec.KSCommand{}
	kscmd.KubeConfigKey = "proj-feature-testing"
	kscmd.Environment = "proj-feature-testing"
	kscmd.Pod = "web-123456"
	kscmd.Stdin = os.Stdin
	kscmd.Stdout = os.Stdout
	kscmd.Stderr = os.Stderr

	spyInitInteractiveH.ExpectsCallCount(t, "Complete", 1)
	spyInitInteractiveH.ExpectsFirstCallArgumentStringSlice(t, "Complete", "argsIn", []string{"ls", "-a", "-l", "-l"})

	spyInitInteractiveH.ExpectsCallCount(t, "Validate", 1)
	spyInitInteractiveH.ExpectsCallCount(t, "Handle", 1)

	spyApi.ExpectsCallCount(t, "SetApiKey", 1)

	spyConfig.ExpectsCallNArgument(t, "Set", 1, "key", config.CpKubeProxyEnabled)
	spyConfig.ExpectsCallNArgument(t, "Set", 1, "value", true)

	spyConfig.ExpectsCallNArgument(t, "Set", 2, "key", config.FlowId)
	spyConfig.ExpectsCallNArgument(t, "Set", 2, "value", "my-flow")

	spyConfig.ExpectsCallNArgument(t, "Set", 3, "key", config.ClusterIdentifier)
	spyConfig.ExpectsCallNArgument(t, "Set", 3, "value", "my-cluster")

	spyLocalExecutor.ExpectsCallCount(t, "StartProcess", 1)
	spyLocalExecutor.ExpectsFirstCallArgument(t, "StartProcess", "kscmd", kscmd)
	spyLocalExecutor.ExpectsFirstCallArgumentStringSlice(t, "StartProcess", "execCmdArgs", []string{"ls", "-a", "-l", "-l"})

	spyKubeCtlInitializer.ExpectsCallCount(t, "Init", 1)

}
