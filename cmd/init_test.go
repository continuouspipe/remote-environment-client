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
	"io/ioutil"
	"k8s.io/client-go/pkg/api/v1"
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
	spyConfig.MockSave(func(configType config.ConfigType) error {
		return nil
	})
	spyApiProvider := spies.NewSpyApiProvider()
	spyApiProvider.MockGetRemoteEnvironmentStatus(func(flowId string, environmentId string) (*cpapi.ApiRemoteEnvironmentStatus, error) {
		r := &cpapi.ApiRemoteEnvironmentStatus{}
		return r, nil
	})

	//get test subject
	handler := &parseSaveTokenInfo{
		spyConfig,
		"some-api-key,remote-env-id,my-project,cp-user,my-branch",
		spyApiProvider}

	handler.Handle()

	//expectations
	spyConfig.ExpectsCallCount(t, "Save", 2)

	spyConfig.ExpectsFirstCallArgument(t, "Set", "key", config.InitStatus)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "value", initStateParseSaveToken)

	spyConfig.ExpectsCallNArgument(t, "Set", 2, "key", config.Username)
	spyConfig.ExpectsCallNArgument(t, "Set", 2, "value", "cp-user")

	spyConfig.ExpectsCallNArgument(t, "Set", 3, "key", config.ApiKey)
	spyConfig.ExpectsCallNArgument(t, "Set", 3, "value", "some-api-key")

	spyConfig.ExpectsCallNArgument(t, "Set", 4, "key", config.FlowId)
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
		case config.FlowId:
			return "837d92hd-19su1d91", nil
		case config.Username:
			return "user-foo", nil
		case config.RemoteName:
			return "origin", nil
		case config.RemoteBranch:
			return "remote-dev-user-foo", nil
		}
		return "", nil
	})
	spyConfig.MockSave(func(configType config.ConfigType) error {
		return nil
	})
	spyConfig.MockSet(func(key string, value interface{}) error {
		return nil
	})

	spyApi := spies.NewSpyApiProvider()
	spyApi.MockGetRemoteEnvironmentStatus(func(flowId string, environmentId string) (*cpapi.ApiRemoteEnvironmentStatus, error) {
		return &cpapi.ApiRemoteEnvironmentStatus{
			Status: cpapi.RemoteEnvironmentTideNotStarted,
		}, nil
	})
	spyApi.MockRemoteEnvironmentBuild(func(remoteEnvId string, gitBranch string) error {
		return nil
	})
	spyApi.MockGetApiEnvironments(func(flowId string) ([]cpapi.ApiEnvironment, error) {
		return []cpapi.ApiEnvironment{}, nil
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
	spyQuestionPrompt := spies.NewSpyQuestionPrompt()
	spyQuestionPrompt.MockRepeatUntilValid(func(question string, isValid func(string) (bool, error)) string {
		return "yes"
	})

	//get test subject
	handler := &triggerBuild{
		spyConfig,
		spyApi,
		spyCommit,
		mocklsRemote,
		spyPush,
		mockRevParse,
		mockWriter,
		spyQuestionPrompt,
	}

	handler.Handle()

	//expectations
	spyConfig.ExpectsCallCount(t, "Save", 1)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "key", config.InitStatus)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "value", initStateTriggerBuild)

	spyApi.ExpectsCallCount(t, "SetApiKey", 1)
	spyApi.ExpectsFirstCallArgument(t, "SetApiKey", "apiKey", "some-api-key")

	spyApi.ExpectsCallCount(t, "GetRemoteEnvironmentStatus", 1)
	spyApi.ExpectsFirstCallArgument(t, "GetRemoteEnvironmentStatus", "flowId", "837d92hd-19su1d91")
	spyApi.ExpectsFirstCallArgument(t, "GetRemoteEnvironmentStatus", "environmentId", "987654321")

	spyPush.ExpectsCallCount(t, "Push", 1)
	spyPush.ExpectsFirstCallArgument(t, "Push", "localBranch", "feature-new")
	spyPush.ExpectsFirstCallArgument(t, "Push", "remoteName", "origin")
	spyPush.ExpectsFirstCallArgument(t, "Push", "remoteBranch", "remote-dev-user-foo")

	spyApi.ExpectsCallCount(t, "RemoteEnvironmentBuild", 1)
	spyApi.ExpectsFirstCallArgument(t, "RemoteEnvironmentBuild", "remoteEnvironmentFlowID", "837d92hd-19su1d91")
	spyApi.ExpectsFirstCallArgument(t, "RemoteEnvironmentBuild", "gitBranch", "remote-dev-user-foo")

	spyApi.ExpectsCallCount(t, "GetApiEnvironments", 1)
	spyApi.ExpectsFirstCallArgument(t, "GetApiEnvironments", "flowId", "837d92hd-19su1d91")
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
		case config.FlowId:
			return "837d92hd-19su1d91", nil
		case config.RemoteBranch:
			return "remote-dev-user-foo", nil
		}
		return "", nil
	})
	spyConfig.MockSave(func(configType config.ConfigType) error {
		return nil
	})
	spyConfig.MockSet(func(key string, value interface{}) error {
		return nil
	})

	//make ticker really quick as is only a test
	mockTicker := time.NewTicker(time.Millisecond * 1)

	spyApi := spies.NewSpyApiProvider()

	spyApi.MockRemoteEnvironmentBuild(func(remoteEnvId string, gitBranch string) error {
		return nil
	})
	spyApi.MockGetApiEnvironments(func(flowId string) ([]cpapi.ApiEnvironment, error) {
		return []cpapi.ApiEnvironment{}, nil
	})
	//mock a response with a status of:
	//RemoteEnvironmentTideFailed the first time
	//RemoteEnvironmentTideRunning the second time
	//RemoteEnvironmentRunning second time
	spyApi.MockGetRemoteEnvironmentStatus(func(flowId string, environmentId string) (*cpapi.ApiRemoteEnvironmentStatus, error) {
		var s string
		callCount := spyApi.CallsCountFor("GetRemoteEnvironmentStatus")

		switch callCount {
		case 1:
			s = cpapi.RemoteEnvironmentTideFailed
		case 2:
			s = cpapi.RemoteEnvironmentTideNotStarted
		case 3:
			s = cpapi.RemoteEnvironmentTideRunning
		case 4:
			s = cpapi.RemoteEnvironmentRunning
		}
		r := &cpapi.ApiRemoteEnvironmentStatus{}
		r.Status = s
		return r, nil
	})

	//get test subject
	handler := waitEnvironmentReady{
		spyConfig,
		spyApi,
		mockTicker,
		ioutil.Discard,
	}
	handler.Handle()

	//expectations
	spyConfig.ExpectsCallCount(t, "Save", 1)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "key", config.InitStatus)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "value", initStateWaitEnvironmentReady)

	spyApi.ExpectsCallCount(t, "GetRemoteEnvironmentStatus", 4)

	spyApi.ExpectsFirstCallArgument(t, "GetRemoteEnvironmentStatus", "flowId", "837d92hd-19su1d91")
	spyApi.ExpectsFirstCallArgument(t, "GetRemoteEnvironmentStatus", "environmentId", "987654321")

	spyApi.ExpectsCallNArgument(t, "GetRemoteEnvironmentStatus", 2, "flowId", "837d92hd-19su1d91")
	spyApi.ExpectsCallNArgument(t, "GetRemoteEnvironmentStatus", 2, "environmentId", "987654321")

	spyApi.ExpectsCallNArgument(t, "GetRemoteEnvironmentStatus", 3, "flowId", "837d92hd-19su1d91")
	spyApi.ExpectsCallNArgument(t, "GetRemoteEnvironmentStatus", 3, "environmentId", "987654321")

	spyApi.ExpectsCallNArgument(t, "GetRemoteEnvironmentStatus", 4, "flowId", "837d92hd-19su1d91")
	spyApi.ExpectsCallNArgument(t, "GetRemoteEnvironmentStatus", 4, "environmentId", "987654321")

	spyApi.ExpectsCallCount(t, "RemoteEnvironmentBuild", 2)
	spyApi.ExpectsFirstCallArgument(t, "RemoteEnvironmentBuild", "remoteEnvironmentFlowID", "837d92hd-19su1d91")
	spyApi.ExpectsFirstCallArgument(t, "RemoteEnvironmentBuild", "gitBranch", "remote-dev-user-foo")

	spyApi.ExpectsCallCount(t, "GetApiEnvironments", 1)
	spyApi.ExpectsFirstCallArgument(t, "GetApiEnvironments", "flowId", "837d92hd-19su1d91")
}

func TestApplyEnvironmentSettings_Handle(t *testing.T) {
	fmt.Println("Running TestApplyEnvironmentSettings_Handle")
	defer fmt.Println("TestApplyEnvironmentSettings_Handle Done")

	//get  mocked dependencies
	spyConfig := spies.NewSpyConfig()
	spyConfig.MockGetString(func(key string) (string, error) {
		switch key {
		case config.ApiKey:
			return "some-api-key", nil
		case config.RemoteEnvironmentId:
			return "987654321", nil
		case config.Username:
			return "user-foo", nil
		case config.FlowId:
			return "837d92hd-19su1d91", nil
		case config.ClusterIdentifier:
			return "the-cluster-one", nil
		case config.CpKubeProxyAddr:
			return "https://kube-proxy-address", nil
		case config.KubeEnvironmentName:
			return "837d92hd-19su1d91-dev-some-user", nil
		}
		return "", nil
	})
	spyConfig.MockConfigFileUsed(func(configType config.ConfigType) (string, error) {
		return "", nil
	})
	spyConfig.MockSave(func(configType config.ConfigType) error {
		return nil
	})
	spyConfig.MockSet(func(key string, value interface{}) error {
		return nil
	})

	spyApi := spies.NewSpyApiProvider()
	spyApi.MockGetRemoteEnvironmentStatus(func(flowId string, environmentId string) (*cpapi.ApiRemoteEnvironmentStatus, error) {
		return &cpapi.ApiRemoteEnvironmentStatus{
			cpapi.RemoteEnvironmentRunning,
			"837d92hd-19su1d91-dev-some-user",
			"the-cluster-one",
			[]cpapi.ApiPublicEndpoint{},
			cpapi.ApiTide{},
		}, nil
	})
	spyApi.MockRemoteEnvironmentBuild(func(remoteEnvId string, gitBranch string) error {
		return nil
	})

	spyKubeCtlInitializer := spies.NewSpyKubeCtlInitializer()
	spyKubeCtlInitializer.MockInit(func(environment string) error {
		return nil
	})

	//get test subject
	handler := &applyEnvironmentSettings{
		spyConfig,
		spyApi,
		spyKubeCtlInitializer,
		ioutil.Discard,
	}
	handler.Handle()

	//expectations
	spyConfig.ExpectsCallCount(t, "Save", 3)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "key", config.InitStatus)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "value", initStateApplyEnvironmentSettings)

	spyApi.ExpectsCallCount(t, "SetApiKey", 1)
	spyApi.ExpectsFirstCallArgument(t, "SetApiKey", "apiKey", "some-api-key")

	spyApi.ExpectsCallCount(t, "GetRemoteEnvironmentStatus", 1)
	spyApi.ExpectsFirstCallArgument(t, "GetRemoteEnvironmentStatus", "flowId", "837d92hd-19su1d91")
	spyApi.ExpectsFirstCallArgument(t, "GetRemoteEnvironmentStatus", "environmentId", "987654321")

	spyConfig.ExpectsCallNArgument(t, "Set", 2, "key", config.ClusterIdentifier)
	spyConfig.ExpectsCallNArgument(t, "Set", 2, "value", "the-cluster-one")
	spyConfig.ExpectsCallNArgument(t, "Set", 3, "key", config.KubeEnvironmentName)
	spyConfig.ExpectsCallNArgument(t, "Set", 3, "value", "837d92hd-19su1d91-dev-some-user")

	spyKubeCtlInitializer.ExpectsCallCount(t, "Init", 1)
}

func TestApplyDefaultService_Handle(t *testing.T) {
	fmt.Println("Running TestApplyDefaultService_Handle")
	defer fmt.Println("TestApplyDefaultService_Handle Done")

	//get mocked dependencies
	spyConfig := spies.NewSpyConfig()
	spyConfig.MockGetString(func(key string) (string, error) {
		switch key {
		case config.KubeEnvironmentName:
			return "837d92hd-19su1d91-dev-some-user", nil
		}
		return "", nil
	})
	spyConfig.MockSave(func(configType config.ConfigType) error {
		return nil
	})
	spyConfig.MockSet(func(key string, value interface{}) error {
		return nil
	})

	spyServiceFinder := spies.NewSpyServiceFinder()
	spyServiceFinder.MockFindAll(func(kubeConfigKey string, environment string) (*v1.ServiceList, error) {
		mockWeb := v1.Service{}
		mockWeb.Name = "app"
		mockDb := v1.Service{}
		mockDb.Name = "db"
		sl := &v1.ServiceList{}
		sl.Items = []v1.Service{mockWeb, mockDb}
		return sl, nil
	})

	spyQuestionPrompt := spies.NewSpyQuestionPrompt()
	spyQuestionPrompt.MockRepeatUntilValid(func(question string, isValid func(string) (bool, error)) string {
		return "1"
	})

	//get test subject
	handler := &applyDefaultService{
		spyConfig,
		spyQuestionPrompt,
		spyServiceFinder,
		ioutil.Discard}

	handler.Handle()

	//expectations
	spyServiceFinder.ExpectsCallCount(t, "FindAll", 1)
	spyServiceFinder.ExpectsFirstCallArgument(t, "FindAll", "kubeConfigKey", "837d92hd-19su1d91-dev-some-user")
	spyServiceFinder.ExpectsFirstCallArgument(t, "FindAll", "environment", "837d92hd-19su1d91-dev-some-user")

	spyQuestionPrompt.ExpectsCallCount(t, "RepeatUntilValid", 1)
	spyQuestionPrompt.ExpectsFirstCallArgument(t, "RepeatUntilValid", "question",
		`Which default container would you like to use?
[0] app
[1] db


Select an option from 0 to 1: `)

	spyConfig.ExpectsCallCount(t, "Save", 2)
	spyConfig.ExpectsCallCount(t, "Set", 2)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "key", config.InitStatus)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "value", initStateApplyDefaultService)
	spyConfig.ExpectsCallNArgument(t, "Set", 2, "key", config.Service)
	spyConfig.ExpectsCallNArgument(t, "Set", 2, "value", "db")
}

func TestInitInteractiveHandler_Handle(t *testing.T) {
	fmt.Println("Running TestInitInteractiveHandler_Handle")
	defer fmt.Println("TestInitInteractiveHandler_Handle Done")

	type expectation func(*testing.T, *spies.SpyConfig, *spies.SpyApiProvider, *spies.SpyQuestionPrompt)
	type scenario struct {
		currentUsername  string
		currentApiKey    string
		insertedUsername string
		insertedApiKey   string
		reset            bool
		expectations     []expectation
	}

	usernameSet := func(t *testing.T, spyConfig *spies.SpyConfig, spyApiProvider *spies.SpyApiProvider, spyQuestionPrompt *spies.SpyQuestionPrompt) {
		spyConfig.ExpectsFirstCallArgument(t, "Set", "key", config.Username)
		spyConfig.ExpectsFirstCallArgument(t, "Set", "value", "foo")
	}
	apiKeySet := func(t *testing.T, spyConfig *spies.SpyConfig, spyApiProvider *spies.SpyApiProvider, spyQuestionPrompt *spies.SpyQuestionPrompt) {
		spyConfig.ExpectsCallNArgument(t, "Set", 2, "key", config.ApiKey)
		spyConfig.ExpectsCallNArgument(t, "Set", 2, "value", "bar")
	}
	configSaved := func(t *testing.T, spyConfig *spies.SpyConfig, spyApiProvider *spies.SpyApiProvider, spyQuestionPrompt *spies.SpyQuestionPrompt) {
		spyConfig.ExpectsCallCount(t, "Save", 1)
		spyConfig.ExpectsFirstCallArgument(t, "Save", "configType", config.GlobalConfigType)
	}

	scenarios := []scenario{
		{
			currentUsername:  "",
			currentApiKey:    "",
			insertedUsername: "foo",
			insertedApiKey:   "bar",
			reset:            false,
			expectations: []expectation{
				usernameSet, apiKeySet, configSaved,
			},
		},
		{
			currentUsername:  "",
			currentApiKey:    "",
			insertedUsername: "foo",
			insertedApiKey:   "bar",
			reset:            true,
			expectations: []expectation{
				usernameSet, apiKeySet, configSaved,
			},
		},
	}

	for _, scenario := range scenarios {
		//get mocked dependencies
		spyConfig := spies.NewSpyConfig()
		spyConfig.MockSet(func(key string, value interface{}) error {
			return nil
		})
		spyConfig.MockSave(func(configType config.ConfigType) error {
			return nil
		})
		spyConfig.MockGetString(func(key string) (string, error) {
			switch key {
			case config.Username:
				return scenario.currentUsername, nil
			case config.ApiKey:
				return scenario.currentApiKey, nil
			}
			return "", nil
		})
		spyApiProvider := spies.NewSpyApiProvider()
		spyApiProvider.MockGetApiUser(func(user string) (*cpapi.ApiUser, error) {
			u := &cpapi.ApiUser{}
			u.Username = scenario.insertedUsername
			return u, nil
		})

		spyQuestionPrompt := spies.NewSpyQuestionPrompt()
		spyQuestionPrompt.MockRepeatIfEmpty(func(question string) string {
			switch question {
			case "Insert your CP Username:":
				return scenario.insertedUsername
			case "Insert your CP Api Key:":
				return scenario.insertedApiKey
			default:
				return "foo"
			}
		})

		//get test subject
		handler := &initInteractiveHandler{}
		handler.api = spyApiProvider
		handler.config = spyConfig
		handler.qp = spyQuestionPrompt
		handler.writer = ioutil.Discard
		handler.reset = false

		//call handler
		err := handler.Handle()
		test.AssertNotError(t, err)

		for _, expectation := range scenario.expectations {
			expectation(t, spyConfig, spyApiProvider, spyQuestionPrompt)
		}
	}

}
