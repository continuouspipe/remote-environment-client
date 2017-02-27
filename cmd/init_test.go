package cmd

import (
	"fmt"
	"testing"

	"encoding/base64"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/test"
	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"time"
)

func TestInitHandler_Complete(t *testing.T) {
	fmt.Println("Running TestInitHandler_Complete")
	defer fmt.Println("TestInitHandler_Complete Done")

	//get mocked dependencies
	spyConfig := spies.NewSpyConfig()
	spyConfig.MockGetString(func(key string) (string, error) {
		return "origin", nil
	})
	spyConfig.MockSet(func(key string, value interface{}) error {
		return nil
	})

	//init is called without providing a token
	handler := &initHandler{}
	handler.config = spyConfig
	err := handler.Complete([]string{""})

	//we expect an error back
	expected := "Invalid token. Please go to continouspipe.io to obtain a valid token"
	test.AssertError(t, expected, err)

	//init is called with a token that is not a base64 string
	err = handler.Complete([]string{"some-token"})

	//we expect an error back
	expected = "Malformed token. Please go to continouspipe.io to obtain a valid token"
	test.AssertError(t, expected, err)

	//init is called providing a valid base64 string
	err = handler.Complete([]string{base64.StdEncoding.EncodeToString([]byte("somethings"))})
	//we don't expect any error
	test.AssertNotError(t, err)
}

func TestInitHandler_Validate(t *testing.T) {
	fmt.Println("Running TestInitHandler_Validate")
	defer fmt.Println("TestInitHandler_Validate Done")

	handler := &initHandler{}

	//Malformed token (not base64 encoded)
	malformedErrorText := "Malformed token. Please go to continouspipe.io to obtain a valid token"

	handler.Complete([]string{"some-token"})
	err := handler.Validate()
	test.AssertError(t, malformedErrorText, err)

	//Token in base64 but with parts missing
	handler.Complete([]string{base64.StdEncoding.EncodeToString([]byte("some-api-key,project,cp-user,my-branch"))})
	err = handler.Validate()
	test.AssertError(t, malformedErrorText, err)

	//Valid token
	handler.Complete([]string{base64.StdEncoding.EncodeToString([]byte("some-api-key,remote-env-id,project,cp-user,my-branch"))})
	err = handler.Validate()
	test.AssertNotError(t, err)
}

func TestInitHandler_Handle_InitStatusAlreadyCompleted(t *testing.T) {
	fmt.Println("Running TestInitHandler_Validate")
	defer fmt.Println("TestInitHandler_Validate Done")

	//get mocked dependencies
	spyConfig := spies.NewSpyConfig()
	spyQuestionPrompt := spies.NewSpyQuestionPrompt()
	spyQuestionPrompt.MockRepeatIfEmpty(func(question string) string {
		return "no"
	})

	//if the initialization status is completed
	spyConfig.MockGetString(func(key string) (string, error) {
		if key == config.InitStatus {
			return initStateCompleted, nil
		}
		return "", nil
	})

	//get test subject
	handler := &initHandler{}
	handler.config = spyConfig
	handler.qp = spyQuestionPrompt
	handler.Handle()

	//Expect that we ask the user if we want to re-initialize
	expectedQuestion := "The environment is already initialized, do you want to re-initialize? (yes/no)"
	spyQuestionPrompt.ExpectsCallCount(t, "RepeatIfEmpty", 1)
	spyQuestionPrompt.ExpectsFirstCallArgument(t, "RepeatIfEmpty", "question", expectedQuestion)
}

func TestParseSaveTokenInfo_Handle(t *testing.T) {
	fmt.Println("Running TestParseSaveTokenInfo_Handle")
	defer fmt.Println("TestParseSaveTokenInfo_Handle Done")

	//get mocked dependencies
	spyConfig := spies.NewSpyConfig()
	spyConfig.MockSet(func(key string, value interface{}) error {
		return nil
	})
	spyConfig.MockSave(func() error {
		return nil
	})
	spyApiProvider := spies.NewSpyApiProvider()
	spyApiProvider.MockGetRemoteEnvironment(func(remoteEnvironmentID string) (*cpapi.ApiRemoteEnvironment, error) {
		r := &cpapi.ApiRemoteEnvironment{}
		return r, nil
	})

	//get test subject
	handler := &parseSaveTokenInfo{
		spyConfig,
		"some-api-key,remote-env-id,my-project,cp-user,my-branch",
		spyApiProvider}

	handler.handle()

	//expectations
	spyConfig.ExpectsCallCount(t, "Save", 2)

	spyConfig.ExpectsFirstCallArgument(t, "Set", "key", config.InitStatus)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "value", initStateParseSaveToken)

	spyConfig.ExpectsCallNArgument(t, "Set", 2, "key", config.Username)
	spyConfig.ExpectsCallNArgument(t, "Set", 2, "value", "cp-user")

	spyConfig.ExpectsCallNArgument(t, "Set", 3, "key", config.ApiKey)
	spyConfig.ExpectsCallNArgument(t, "Set", 3, "value", "some-api-key")

	spyConfig.ExpectsCallNArgument(t, "Set", 4, "key", config.Project)
	spyConfig.ExpectsCallNArgument(t, "Set", 4, "value", "my-project")

	spyConfig.ExpectsCallNArgument(t, "Set", 5, "key", config.RemoteBranch)
	spyConfig.ExpectsCallNArgument(t, "Set", 5, "value", "my-branch")

	spyConfig.ExpectsCallNArgument(t, "Set", 6, "key", config.RemoteEnvironmentId)
	spyConfig.ExpectsCallNArgument(t, "Set", 6, "value", "remote-env-id")
}

func TestTriggerBuild_Handle(t *testing.T) {
	fmt.Println("Running TestTriggerBuild_Handle")
	defer fmt.Println("TestTriggerBuild_Handle Done")

	//get mocked dependencies
	spyConfig := spies.NewSpyConfig()
	spyConfig.MockGetString(func(key string) (string, error) {
		switch key {
		case config.ApiKey:
			return "some-api-key", nil
		case config.RemoteEnvironmentId:
			return "987654321", nil
		case config.Username:
			return "user-foo", nil
		case config.RemoteName:
			return "origin", nil
		case config.RemoteBranch:
			return "remote-dev-user-foo", nil
		}
		return "", nil
	})
	spyConfig.MockSave(func() error {
		return nil
	})
	spyConfig.MockSet(func(key string, value interface{}) error {
		return nil
	})

	spyApi := spies.NewSpyApiProvider()
	spyApi.MockGetRemoteEnvironment(func(remoteEnvId string) (*cpapi.ApiRemoteEnvironment, error) {
		return &cpapi.ApiRemoteEnvironment{
			Status: cpapi.RemoteEnvironmentStatusNotStarted,
		}, nil
	})
	spyApi.MockRemoteEnvironmentBuild(func(remoteEnvId string) error {
		return nil
	})

	mocklsRemote := mocks.NewMockLsRemote()
	mocklsRemote.MockGetList(func(remoteName string, remoteBranch string) (string, error) {
		return "", nil
	})
	mockRevParse := mocks.NewMockRevParse()
	mockRevParse.MockGetLocalBranchName(func() (string, error) {
		return "feature-new", nil
	})
	spyCommit := spies.NewSpyCommit()
	spyPush := spies.NewSpyPush()
	spyPush.MockPush(func() (string, error) {
		return "", nil
	})

	mockWriter := spies.NewSpyWriter()
	mockWriter.MockWrite(func(p []byte) (n int, err error) {
		return 100, nil
	})

	//get test subject
	handler := &triggerBuild{
		spyConfig,
		spyApi,
		spyCommit,
		mocklsRemote,
		spyPush,
		mockRevParse,
		mockWriter}

	handler.handle()

	//expectations
	spyConfig.ExpectsCallCount(t, "Save", 1)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "key", config.InitStatus)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "value", initStateTriggerBuild)

	spyApi.ExpectsCallCount(t, "SetApiKey", 1)
	spyApi.ExpectsFirstCallArgument(t, "SetApiKey", "apiKey", "some-api-key")

	spyApi.ExpectsCallCount(t, "GetRemoteEnvironment", 1)
	spyApi.ExpectsFirstCallArgument(t, "GetRemoteEnvironment", "remoteEnvironmentID", "987654321")

	spyPush.ExpectsCallCount(t, "Push", 1)
	spyPush.ExpectsFirstCallArgument(t, "Push", "localBranch", "feature-new")
	spyPush.ExpectsFirstCallArgument(t, "Push", "remoteName", "origin")
	spyPush.ExpectsFirstCallArgument(t, "Push", "remoteBranch", "remote-dev-user-foo")
}

func TestWaitEnvironmentReady_Handle(t *testing.T) {
	fmt.Println("Running TestWaitEnvironmentReady_Handle")
	defer fmt.Println("TestWaitEnvironmentReady_Handle Done")

	//get mocked dependencies
	spyConfig := spies.NewSpyConfig()
	spyConfig.MockGetString(func(key string) (string, error) {
		switch key {
		case config.ApiKey:
			return "some-api-key", nil
		case config.RemoteEnvironmentId:
			return "987654321", nil
		}
		return "", nil
	})
	spyConfig.MockSave(func() error {
		return nil
	})
	spyConfig.MockSet(func(key string, value interface{}) error {
		return nil
	})

	//make ticker really quick as is only a test
	mockTicker := time.NewTicker(time.Millisecond * 1)

	spyApi := spies.NewSpyApiProvider()

	spyApi.MockRemoteEnvironmentBuild(func(remoteEnvId string) error {
		return nil
	})

	//mock a response with a status of:
	//RemoteEnvironmentStatusNotStarted the first time
	//RemoteEnvironmentStatusBuilding the second time
	//RemoteEnvironmentStatusOk second time
	spyApi.MockGetRemoteEnvironment(func(remoteEnvironmentID string) (*cpapi.ApiRemoteEnvironment, error) {
		var s string
		callCount := spyApi.CallsCountFor("GetRemoteEnvironment")

		switch callCount {
		case 1:
			s = cpapi.RemoteEnvironmentStatusNotStarted
		case 2:
			s = cpapi.RemoteEnvironmentStatusBuilding
		case 3:
			s = cpapi.RemoteEnvironmentStatusOk
		}
		r := &cpapi.ApiRemoteEnvironment{}
		r.Status = s
		return r, nil
	})

	//get test subject
	handler := waitEnvironmentReady{
		spyConfig,
		spyApi,
		mockTicker,
	}
	handler.handle()

	//expectations
	spyConfig.ExpectsCallCount(t, "Save", 1)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "key", config.InitStatus)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "value", initStateWaitEnvironmentReady)

	spyApi.ExpectsCallCount(t, "GetRemoteEnvironment", 3)
	spyApi.ExpectsFirstCallArgument(t, "GetRemoteEnvironment", "remoteEnvironmentID", "987654321")
	spyApi.ExpectsCallNArgument(t, "GetRemoteEnvironment", 2, "remoteEnvironmentID", "987654321")
	spyApi.ExpectsCallNArgument(t, "GetRemoteEnvironment", 3, "remoteEnvironmentID", "987654321")

	spyApi.ExpectsCallCount(t, "RemoteEnvironmentBuild", 1)
	spyApi.ExpectsFirstCallArgument(t, "RemoteEnvironmentBuild", "remoteEnvironmentID", "987654321")
}
