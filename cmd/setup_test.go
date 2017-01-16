package cmd

import (
	"testing"

	envconfig "github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/test"
)

func TestUserApplicationSettingsAreStored(t *testing.T) {

	settings := &envconfig.ApplicationSettings{
		ProjectKey:            "my-project",
		RemoteBranch:          "feature/MYPROJ-312-initial-commit",
		RemoteName:            "",
		DefaultService:        "web",
		ClusterIp:             "127.0.0.1",
		Username:              "root",
		Password:              "123456",
		AnybarPort:            "6542",
		KeenWriteKey:          "sk29dj22d882",
		KeenProjectId:         "cc3d902idi01",
		KeenEventCollection:   "event-collection",
		Environment:           "",
	}

	expectedSettings := *settings

	//this is the default expected value for RemoteName
	expectedSettings.RemoteName = "origin"

	//we expect / to be converted to - and namespace being a concatenation of ProjectKey and RemoteBranch
	expectedSettings.Environment = "my-project-feature-MYPROJ-312-initial-commit"

	mockedQuestionPrompt := &MockQuestionPrompt{settings}
	mockedYamlWriter := &MockYamlWriter{
		test:             t,
		expectedSettings: &expectedSettings,
	}

	setupHandle := &SetupHandle{}
	setupHandle.storeUserSettings(mockedQuestionPrompt, mockedYamlWriter)
}

type MockQuestionPrompt struct {
	testSettings *envconfig.ApplicationSettings
}

func (qp MockQuestionPrompt) ReadString(q string) string {
	questions := [11]struct {
		question, answer string
	}{
		{"What is your Continuous Pipe project key?", qp.testSettings.ProjectKey},
		{"What is the name of the Git branch you are using for your remote environment?", qp.testSettings.RemoteBranch},
		{"What is your github remote name? (defaults to: origin)", qp.testSettings.RemoteName},
		{"What is the default container for the watch, bash, fetch and resync commands?", qp.testSettings.DefaultService},
		{"What is the IP of the cluster?", qp.testSettings.ClusterIp},
		{"What is the cluster username?", qp.testSettings.Username},
		{"What is the cluster password?", qp.testSettings.Password},
		{"If you want to use AnyBar, please provide a port number e.g 1738 ?", qp.testSettings.AnybarPort},
		{"What is your keen.io write key? (Optional, only needed if you want to record usage stats)", qp.testSettings.KeenWriteKey},
		{"What is your keen.io project id? (Optional, only needed if you want to record usage stats)", qp.testSettings.KeenProjectId},
		{"What is your keen.io event collection?  (Optional, only needed if you want to record usage stats)", qp.testSettings.KeenEventCollection},
	}
	for _, v := range questions {
		if q == v.question {
			return v.answer
		}
	}
	return ""
}

func (qp MockQuestionPrompt) ApplyDefault(question string, predef string) string {
	return predef
}

func (qp MockQuestionPrompt) RepeatIfEmpty(question string) string {
	return qp.ReadString(question)
}

type MockYamlWriter struct {
	test             *testing.T
	expectedSettings *envconfig.ApplicationSettings
}

func (w MockYamlWriter) Save(c *envconfig.ApplicationSettings) bool {
	test.AssertSame(w.test, w.expectedSettings.ProjectKey, c.ProjectKey)
	test.AssertSame(w.test, w.expectedSettings.RemoteBranch, c.RemoteBranch)
	test.AssertSame(w.test, w.expectedSettings.RemoteName, c.RemoteName)
	test.AssertSame(w.test, w.expectedSettings.DefaultService, c.DefaultService)
	test.AssertSame(w.test, w.expectedSettings.ClusterIp, c.ClusterIp)
	test.AssertSame(w.test, w.expectedSettings.Username, c.Username)
	test.AssertSame(w.test, w.expectedSettings.Password, c.Password)
	test.AssertSame(w.test, w.expectedSettings.AnybarPort, c.AnybarPort)
	test.AssertSame(w.test, w.expectedSettings.KeenWriteKey, c.KeenWriteKey)
	test.AssertSame(w.test, w.expectedSettings.KeenProjectId, c.KeenProjectId)
	test.AssertSame(w.test, w.expectedSettings.KeenEventCollection, c.KeenEventCollection)
	test.AssertSame(w.test, w.expectedSettings.Environment, c.Environment)
	return true
}
