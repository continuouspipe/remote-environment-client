//TODO: Refactor test to testify framework https://github.com/stretchr/testify
package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"testing"
	"github.com/stretchr/testify/mock"
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
	buildHandle := BuildHandle{}
	buildHandle.triggerBuild = spyTriggerBuild
	buildHandle.waitForEnvironmentReady = spyWaitForEnvironmentReadyState
	buildHandle.stdout = mockStdout
	buildHandle.config = spyConfig
	buildHandle.api = spyApiProvider
	buildHandle.Handle()

	//expectations
	spyTriggerBuild.ExpectsCallCount(t, "Handle", 1)
	spyWaitForEnvironmentReadyState.ExpectsCallCount(t, "Handle", 1)
}
