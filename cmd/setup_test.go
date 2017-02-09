package cmd

import (
	"testing"

	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/test"
	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
)

func TestUserApplicationSettingsAreStored(t *testing.T) {
	fmt.Println("Running TestUserApplicationSettingsAreStored")
	defer fmt.Println("TestUserApplicationSettingsAreStored Done")

	//get mocked dependencies
	mockedQuestionPrompt := mocks.NewMockQuestionPrompt()
	spyYamlWriter := spies.NewSpyYamlWriter()
	spyYamlWriter.MockSave(func() bool {
		return true
	})

	//test subject called
	setupHandle := &SetupHandle{}
	setupHandle.storeUserSettings(mockedQuestionPrompt, spyYamlWriter)

	//expectations
	if spyYamlWriter.CallsCountFor("Save") != 1 {
		t.Error("Expected Save to be called only once")
	}

	expectedSettings := &config.ApplicationSettings{
		ProjectKey:   "my-project",
		RemoteBranch: "feature/myproj-312-initial-commit",
		//this is the default expected value for RemoteName
		RemoteName:          "origin",
		DefaultService:      "web",
		ClusterIp:           "127.0.0.1",
		Username:            "root",
		Password:            "2e9fik2s9-fds903",
		AnybarPort:          "6542",
		KeenWriteKey:        "sk29dj22d882",
		KeenProjectId:       "cc3d902idi01",
		KeenEventCollection: "event-collection",
		//we expect / to be converted to - and namespace being a concatenation of ProjectKey and RemoteBranch
		Environment: "my-project-feature-MYPROJ-312-initial-commit",
	}

	spyYamlWriter.ExpectsCallCount(t, "Save", 1)

	firstCall := spyYamlWriter.FirstCallsFor("Save")
	if actualSettings, ok := firstCall.Arguments["settings"].(*config.ApplicationSettings); ok {
		test.AssertDeepEqual(t, expectedSettings, actualSettings)
	} else {
		t.Fatalf("Expected saved settings to be *config.ApplicationSettings, given %T", firstCall.Arguments["settings"])
	}
}
