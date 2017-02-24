package cmd

import (
	"testing"

	"fmt"
	"os"

	kexec "github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/test"
	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"k8s.io/client-go/pkg/api/v1"
)

func TestCommandsAreSpawned(t *testing.T) {
	fmt.Println("Running TestCommandsAreSpawned")
	defer fmt.Println("TestCommandsAreSpawned Done")

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
	spyLocalExecutor.MockCommandExec(func() (string, error) {
		return "some results back..", nil
	})

	//test subject called
	handler := &ExecHandle{}
	handler.Environment = "proj-feature-testing"
	handler.Service = "web"
	out, _ := handler.Handle([]string{"ls", "-a", "-l", "-l"}, mockPodsFinder, mockPodFilter, spyLocalExecutor)

	//expectations
	test.AssertSame(t, "some results back..", out)

	kscmd := kexec.KSCommand{}
	kscmd.KubeConfigKey = "proj-feature-testing"
	kscmd.Environment = "proj-feature-testing"
	kscmd.Pod = "web-123456"
	kscmd.Stdin = os.Stdin
	kscmd.Stdout = os.Stdout
	kscmd.Stderr = os.Stderr

	spyLocalExecutor.ExpectsCallCount(t, "CommandExec", 1)
	spyLocalExecutor.ExpectsFirstCallArgument(t, "CommandExec", "kscmd", kscmd)
	spyLocalExecutor.ExpectsFirstCallArgumentStringSlice(t, "CommandExec", "execCmdArgs", []string{"ls", "-a", "-l", "-l"})
}
