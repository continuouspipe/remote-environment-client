package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"testing"
)

func TestRemoteBranchNotPresent(t *testing.T) {
	fmt.Println("Running TestRemoteBranchNotPresent")
	defer fmt.Println("TestRemoteBranchNotPresent Done")

	//get mocked dependencies
	mockStdout := spies.NewSpyWriter()
	mockStdout.MockWrite(func(p []byte) (n int, err error) {
		return 100, nil
	})
	spyTriggerBuild := spies.NewSpyInitState()
	spyTriggerBuild.MockHandle(func() error {
		return nil
	})
	spyWaitForEnvironmentReadyState := spies.NewSpyInitState()
	spyWaitForEnvironmentReadyState.MockHandle(func() error {
		return nil
	})

	//test subject called
	buildHandle := BuildHandle{}
	buildHandle.triggerBuild = spyTriggerBuild
	buildHandle.waitForEnvironmentReady = spyWaitForEnvironmentReadyState
	buildHandle.stdout = mockStdout
	buildHandle.Handle()

	//expectations
	spyTriggerBuild.ExpectsCallCount(t, "Handle", 1)
	spyWaitForEnvironmentReadyState.ExpectsCallCount(t, "Handle", 1)
}
