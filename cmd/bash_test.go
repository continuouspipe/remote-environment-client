package cmd

import (
	"fmt"
	"os"
	"testing"

	kexec "github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"k8s.io/client-go/pkg/api/v1"
)

func TestSysCallIsCalledToOpenBashSession(t *testing.T) {
	fmt.Println("Running TestSysCallIsCalledToOpenBashSession")
	defer fmt.Println("TestSysCallIsCalledToOpenBashSession Done")

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
	spyLocalExecutor := spies.NewSpyLocalExecutor()
	spyLocalExecutor.MockStartProcess(func() error {
		return nil
	})
	spyKubeCtlInitializer := spies.NewSpyKubeCtlInitializer()
	spyKubeCtlInitializer.MockInit(func(environment string) error {
		return nil
	})

	//test subject called
	handler := &BashHandle{}
	handler.kubeCtlInit = spyKubeCtlInitializer
	handler.Environment = "proj-feature-testing"
	handler.Service = "web"
	handler.Handle([]string{}, mockPodsFinder, mockPodFilter, spyLocalExecutor)

	kscmd := kexec.KSCommand{}
	kscmd.KubeConfigKey = "proj-feature-testing"
	kscmd.Environment = "proj-feature-testing"
	kscmd.Pod = "web-123456"
	kscmd.Stdin = os.Stdin
	kscmd.Stdout = os.Stdout
	kscmd.Stderr = os.Stderr

	//expectations
	spyLocalExecutor.ExpectsCallCount(t, "StartProcess", 1)
	spyLocalExecutor.ExpectsFirstCallArgument(t, "StartProcess", "kscmd", kscmd)
	spyLocalExecutor.ExpectsFirstCallArgumentStringSlice(t, "StartProcess", "execCmdArgs", []string{"/bin/bash"})

	spyKubeCtlInitializer.ExpectsCallCount(t, "Init", 1)
}
