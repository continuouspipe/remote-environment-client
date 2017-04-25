package cmd

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"io/ioutil"
	"testing"
)

func TestRemoteBranchNotPresent(t *testing.T) {
	//get mocked dependencies
	triggerBuild := mocks.NewMockInitState()
	waitForEnvironmentReadyState := mocks.NewMockInitState()
	apiProvider := mocks.NewMockCpApiProvider()
	spyConfig := mocks.NewSpyConfig()

	//set expectations
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
	apiProvider.On("GetRemoteEnvironmentStatus", "837d92hd-19su1d91", "987654321").Return(r, nil)
	apiProvider.On("SetApiKey", "some-api-key")

	spyConfig.
		On("Set", config.InitStatus, initStateCompleted).Return(nil).
		On("Save", config.AllConfigTypes).Return(nil).
		On("GetString", config.ApiKey).Return("some-api-key", nil).
		On("GetString", config.RemoteEnvironmentId).Return("987654321", nil).
		On("GetString", config.FlowId).Return("837d92hd-19su1d91", nil)

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
