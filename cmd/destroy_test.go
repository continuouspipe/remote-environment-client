package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"testing"
	"github.com/continuouspipe/remote-environment-client/errors"
)

func TestDestroyHandle_Handle(t *testing.T) {
	fmt.Println("Running TestDestroyHandle_Handle")
	defer fmt.Println("TestDestroyHandle_Handle Done")

	//get mocked dependencies
	mockStdout := spies.NewSpyWriter()
	mockStdout.MockWrite(func(p []byte) (n int, err error) {
		return 100, nil
	})
	spyConfig := spies.NewSpyConfig()
	spyConfig.MockGetString(func(key string) (string, error) {
		switch key {
		case config.ApiKey:
			return "some-api-key", nil
		case config.ClusterIdentifier:
			return "my-cluster", nil
		case config.FlowId:
			return "837d92hd-19su1d91", nil
		case config.KubeEnvironmentName:
			return "29fsdfjk2d9sj-sadfj2-32342-remote-dev-user-foo", nil
		case config.RemoteName:
			return "origin", nil
		case config.RemoteBranch:
			return "remote-dev-user-foo", nil
		case config.RemoteEnvironmentId:
			return "remote-env-id", nil
		}
		return "", nil
	})
	spyApiProvider := spies.NewSpyApiProvider()
	spyApiProvider.MockGetRemoteEnvironmentStatus(func(flowId string, environmentId string) (*cpapi.ApiRemoteEnvironmentStatus, *errors.ErrorList) {
		r := &cpapi.ApiRemoteEnvironmentStatus{}
		return r, nil
	})
	spyApiProvider.MockCancelRunningTide(func(flowId string, gitBranch string) error {
		return nil
	})
	spyApiProvider.MockRemoteEnvironmentDestroy(func(flowId string, environment string, cluster string) error {
		return nil
	})
	spyApiProvider.MockRemoteDevelopmentEnvironmentDestroy(func(flowId string, remoteEnvironmentId string) error {
		return nil
	})
	mockLsRemote := mocks.NewMockLsRemote()
	mockLsRemote.MockGetList(func(remoteName string, remoteBranch string) (string, error) {
		return "origin/remote-dev-user-foo", nil
	})
	spyPush := spies.NewSpyPush()
	spyPush.MockDeleteRemote(func(remoteName string, remoteBranch string) (string, error) {
		return "", nil
	})
	spyQuestionPrompt := spies.NewSpyQuestionPrompt()
	spyQuestionPrompt.MockRepeatUntilValid(func(question string, isValid func(string) (bool, error)) string {
		return "yes"
	})

	//test subject called
	handler := NewDestroyHandle()
	handler.api = spyApiProvider
	handler.config = spyConfig
	handler.lsRemote = mockLsRemote
	handler.push = spyPush
	handler.qp = spyQuestionPrompt
	handler.stdout = mockStdout
	handler.Handle()

	//expectations

	spyQuestionPrompt.ExpectsCallCount(t, "RepeatUntilValid", 1)
	spyQuestionPrompt.ExpectsFirstCallArgument(t, "RepeatUntilValid", "question", "This will delete the remote git branch and remote environment, do you want to proceed (yes/no)")

	spyApiProvider.ExpectsCallCount(t, "SetApiKey", 1)
	spyApiProvider.ExpectsFirstCallArgument(t, "SetApiKey", "apiKey", "some-api-key")

	spyApiProvider.ExpectsCallCount(t, "CancelRunningTide", 1)
	spyApiProvider.ExpectsFirstCallArgument(t, "CancelRunningTide", "flowId", "837d92hd-19su1d91")
	spyApiProvider.ExpectsFirstCallArgument(t, "CancelRunningTide", "remoteEnvironmentId", "remote-env-id")

	spyPush.ExpectsCallCount(t, "DeleteRemote", 1)
	spyPush.ExpectsFirstCallArgument(t, "DeleteRemote", "remoteName", "origin")
	spyPush.ExpectsFirstCallArgument(t, "DeleteRemote", "gitBranch", "remote-dev-user-foo")
}
