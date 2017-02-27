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
	spyApi.MockRemoteEnvironmentBuild(func(remoteEnvId string, gitBranch string) error {
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

	spyApi.ExpectsCallCount(t, "RemoteEnvironmentBuild", 1)
	spyApi.ExpectsFirstCallArgument(t, "RemoteEnvironmentBuild", "remoteEnvironmentID", "987654321")
	spyApi.ExpectsFirstCallArgument(t, "RemoteEnvironmentBuild", "gitBranch", "remote-dev-user-foo")
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

	//make ticker really quick as is only a test
	mockTicker := time.NewTicker(time.Millisecond * 1)

	spyApi := spies.NewSpyApiProvider()

	spyApi.MockRemoteEnvironmentBuild(func(remoteEnvId string, gitBranch string) error {
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
	spyApi.ExpectsFirstCallArgument(t, "RemoteEnvironmentBuild", "gitBranch", "remote-dev-user-foo")
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
		case config.Project:
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
	spyConfig.MockSave(func() error {
		return nil
	})
	spyConfig.MockSet(func(key string, value interface{}) error {
		return nil
	})

	spyApi := spies.NewSpyApiProvider()
	spyApi.MockGetRemoteEnvironment(func(remoteEnvId string) (*cpapi.ApiRemoteEnvironment, error) {
		return &cpapi.ApiRemoteEnvironment{
			cpapi.RemoteEnvironmentStatusOk,
			"2017-01-01 10:20",
			"2233445566",
			"837d92hd-19su1d91-dev-some-user",
			"the-cluster-one",
			"59846",
			"keen-id-123",
			"keen-write-key-456",
			"proj-events",
		}, nil
	})
	spyApi.MockRemoteEnvironmentBuild(func(remoteEnvId string, gitBranch string) error {
		return nil
	})

	spyKubeCtlConfig := spies.NewSpyKubeCtlConfigProvider()
	spyKubeCtlConfig.MockConfigSetAuthInfo(func(environment string, username string, password string) (string, error) {
		return "", nil
	})
	spyKubeCtlConfig.MockConfigSetCluster(func(environment string, clusterIp string, teamName string, clusterIdentifier string) (string, error) {
		return "", nil
	})
	spyKubeCtlConfig.MockConfigSetContext(func(environment string, username string) (string, error) {
		return "", nil
	})
	spyClusterInfoProvider := spies.NewSpyKubeCtlClusterInfoProvider()
	spyClusterInfoProvider.MockClusterInfo(func(kubeConfigKey string) (string, error) {
		return "", nil
	})

	//get test subject
	handler := &applyEnvironmentSettings{
		spyConfig,
		spyApi,
		spyKubeCtlConfig,
		spyClusterInfoProvider,
		ioutil.Discard,
	}
	handler.handle()

	//expectations
	spyConfig.ExpectsCallCount(t, "Save", 3)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "key", config.InitStatus)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "value", initStateApplyEnvironmentSettings)

	spyApi.ExpectsCallCount(t, "SetApiKey", 1)
	spyApi.ExpectsFirstCallArgument(t, "SetApiKey", "apiKey", "some-api-key")

	spyApi.ExpectsCallCount(t, "GetRemoteEnvironment", 1)
	spyApi.ExpectsFirstCallArgument(t, "GetRemoteEnvironment", "remoteEnvironmentID", "987654321")

	spyConfig.ExpectsCallNArgument(t, "Set", 2, "key", config.RemoteEnvironmentConfigModifiedAt)
	spyConfig.ExpectsCallNArgument(t, "Set", 2, "value", "2017-01-01 10:20")
	spyConfig.ExpectsCallNArgument(t, "Set", 3, "key", config.ClusterIdentifier)
	spyConfig.ExpectsCallNArgument(t, "Set", 3, "value", "the-cluster-one")
	spyConfig.ExpectsCallNArgument(t, "Set", 4, "key", config.AnybarPort)
	spyConfig.ExpectsCallNArgument(t, "Set", 4, "value", "59846")
	spyConfig.ExpectsCallNArgument(t, "Set", 5, "key", config.KubeEnvironmentName)
	spyConfig.ExpectsCallNArgument(t, "Set", 5, "value", "837d92hd-19su1d91-dev-some-user")
	spyConfig.ExpectsCallNArgument(t, "Set", 6, "key", config.KeenEventCollection)
	spyConfig.ExpectsCallNArgument(t, "Set", 6, "value", "proj-events")
	spyConfig.ExpectsCallNArgument(t, "Set", 7, "key", config.KeenProjectId)
	spyConfig.ExpectsCallNArgument(t, "Set", 7, "value", "keen-id-123")
	spyConfig.ExpectsCallNArgument(t, "Set", 8, "key", config.KeenWriteKey)
	spyConfig.ExpectsCallNArgument(t, "Set", 8, "value", "keen-write-key-456")

	spyKubeCtlConfig.ExpectsCallCount(t, "ConfigSetAuthInfo", 1)
	spyKubeCtlConfig.ExpectsFirstCallArgument(t, "ConfigSetAuthInfo", "environment", "837d92hd-19su1d91-dev-some-user")
	spyKubeCtlConfig.ExpectsFirstCallArgument(t, "ConfigSetAuthInfo", "username", "user-foo")
	spyKubeCtlConfig.ExpectsFirstCallArgument(t, "ConfigSetAuthInfo", "password", "some-api-key")

	spyKubeCtlConfig.ExpectsCallCount(t, "ConfigSetCluster", 1)
	spyKubeCtlConfig.ExpectsFirstCallArgument(t, "ConfigSetCluster", "environment", "837d92hd-19su1d91-dev-some-user")
	spyKubeCtlConfig.ExpectsFirstCallArgument(t, "ConfigSetCluster", "clusterIp", "https://kube-proxy-address")
	spyKubeCtlConfig.ExpectsFirstCallArgument(t, "ConfigSetCluster", "teamName", "837d92hd-19su1d91")
	spyKubeCtlConfig.ExpectsFirstCallArgument(t, "ConfigSetCluster", "clusterIdentifier", "the-cluster-one")

	spyKubeCtlConfig.ExpectsCallCount(t, "ConfigSetContext", 1)
	spyKubeCtlConfig.ExpectsFirstCallArgument(t, "ConfigSetContext", "environment", "837d92hd-19su1d91-dev-some-user")
	spyKubeCtlConfig.ExpectsFirstCallArgument(t, "ConfigSetContext", "username", "user-foo")

	spyClusterInfoProvider.ExpectsCallCount(t, "ClusterInfo", 1)
	spyClusterInfoProvider.ExpectsFirstCallArgument(t, "ClusterInfo", "kubeConfigKey", "837d92hd-19su1d91-dev-some-user")
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
	spyConfig.MockSave(func() error {
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
		spyServiceFinder}

	handler.handle()

	//expectations
	spyServiceFinder.ExpectsCallCount(t, "FindAll", 1)
	spyServiceFinder.ExpectsFirstCallArgument(t, "FindAll", "kubeConfigKey", "837d92hd-19su1d91-dev-some-user")
	spyServiceFinder.ExpectsFirstCallArgument(t, "FindAll", "environment", "837d92hd-19su1d91-dev-some-user")

	spyQuestionPrompt.ExpectsCallCount(t, "RepeatUntilValid", 1)
	spyQuestionPrompt.ExpectsFirstCallArgument(t, "RepeatUntilValid", "question",
		"You have 2 services available in you remote environment.\n"+
			"Which one you want to be the default service to be used for commands like: watch, fetch, bash and exec?\n"+
			"Choose an option [0-1]\n\n"+
			"[0] app\n"+
			"[1] db\n")

	spyConfig.ExpectsCallCount(t, "Save", 2)
	spyConfig.ExpectsCallCount(t, "Set", 2)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "key", config.InitStatus)
	spyConfig.ExpectsFirstCallArgument(t, "Set", "value", initStateApplyDefaultService)
	spyConfig.ExpectsCallNArgument(t, "Set", 2, "key", config.Service)
	spyConfig.ExpectsCallNArgument(t, "Set", 2, "value", "db")
}
