package cmd

import (
	"io/ioutil"
	"testing"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/test/mocks"
)

func TestRemoteBranchNotPresent(t *testing.T) {
	//get mocked dependencies
	triggerBuild := mocks.NewMockInitState()
	waitForEnvironmentReadyState := mocks.NewMockInitState()
	apiProvider := mocks.NewMockCpAPIProvider()
	spyConfig := mocks.NewSpyConfig()

	//set expectations
	r := &cpapi.APIRemoteEnvironmentStatus{}
	r.PublicEndpoints = []cpapi.APIPublicEndpoint{
		{
			Address: "10.0.0.0",
			Name:    "web",
			Ports: []cpapi.APIPublicEndpointPort{
				{
					Number:   80,
					Protocol: "tcp",
				},
			},
		},
	}
	apiProvider.On("GetRemoteEnvironmentStatus", "837d92hd-19su1d91", "987654321").Return(r, nil)
	apiProvider.On("SetAPIKey", "some-api-key")

	spyConfig.
		On("Set", config.InitStatus, initStateCompleted).Return(nil).
		On("Save", config.AllConfigTypes).Return(nil).
		On("GetStringQ", config.ApiKey).Return("some-api-key", nil).
		On("GetStringQ", config.RemoteEnvironmentId).Return("987654321", nil).
		On("GetStringQ", config.FlowId).Return("837d92hd-19su1d91", nil)

	triggerBuild.On("Handle").Return(nil)
	waitForEnvironmentReadyState.On("Handle").Return(nil)

	//call the code we are testing
	buildHandle := BuildHandle{}
	buildHandle.triggerBuild = triggerBuild
	buildHandle.waitForEnvironmentReady = waitForEnvironmentReadyState
	buildHandle.stdout = ioutil.Discard
	buildHandle.config = spyConfig
	buildHandle.api = apiProvider
	buildHandle.Handle()

	// assert that the expectations were met
	apiProvider.AssertExpectations(t)
	triggerBuild.AssertExpectations(t)
	waitForEnvironmentReadyState.AssertExpectations(t)
	spyConfig.AssertExpectations(t)

	spyConfig.AssertNumberOfCalls(t, "Set", 1)
	triggerBuild.AssertNumberOfCalls(t, "Handle", 1)
	waitForEnvironmentReadyState.AssertNumberOfCalls(t, "Handle", 1)
}
