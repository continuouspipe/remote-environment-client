package cmd

import (
	"testing"
	envconfig "github.com/continuouspipe/remote-environment-client/config"
)

func TestUserDefinedConfigurationParametersAreStoredOnDisk(t *testing.T) {

	settings := &envconfig.ApplicationSettings{
		ProjectKey:          "my-project",
		RemoteBranch:        "feature/MYPROJ-312-initial-commit",
		RemoteName:          "origin",
		DefaultContainer:    "default-container",
		ClusterIp:           "127.0.0.1",
		Username:            "root",
		Password:            "123456",
		AnybarPort:          "6542",
		KeenWriteKey:        "sk29dj22d882",
		KeenProjectId:       "cc3d902idi01",
		KeenEventCollection: "event-collection",
		Namespace:           "",
	}

	expectedSettings := *settings;
	expectedSettings.Namespace = "my-project-feature-MYPROJ-312-initial-commit"

	mockedQuestionPrompt := &MockQuestionPrompt{settings}
	mockedYamlWriter := &MockYamlWriter{
		test:             t,
		expectedSettings: &expectedSettings,
	}

	runSetupCmd(mockedQuestionPrompt, mockedYamlWriter)
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
		{"What is the default container for the watch, bash, fetch and resync commands? (Optional)", qp.testSettings.DefaultContainer},
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
	return qp.ReadString(question)
}

func (qp MockQuestionPrompt) RepeatIfEmpty(question string) string {
	return qp.ReadString(question)
}

type MockYamlWriter struct {
	test             *testing.T
	expectedSettings *envconfig.ApplicationSettings
}

func (w MockYamlWriter) Save(c *envconfig.ApplicationSettings) bool {
	assertSame(w.test, w.expectedSettings.ProjectKey, c.ProjectKey)
	assertSame(w.test, w.expectedSettings.RemoteBranch, c.RemoteBranch)
	assertSame(w.test, w.expectedSettings.RemoteName, c.RemoteName)
	assertSame(w.test, w.expectedSettings.DefaultContainer, c.DefaultContainer)
	assertSame(w.test, w.expectedSettings.ClusterIp, c.ClusterIp)
	assertSame(w.test, w.expectedSettings.Username, c.Username)
	assertSame(w.test, w.expectedSettings.Password, c.Password)
	assertSame(w.test, w.expectedSettings.AnybarPort, c.AnybarPort)
	assertSame(w.test, w.expectedSettings.KeenWriteKey, c.KeenWriteKey)
	assertSame(w.test, w.expectedSettings.KeenProjectId, c.KeenProjectId)
	assertSame(w.test, w.expectedSettings.KeenEventCollection, c.KeenEventCollection)
	assertSame(w.test, w.expectedSettings.Namespace, c.Namespace)
	return true
}

func assertSame(t *testing.T, expected string, given string) {
	if given != expected {
		t.Errorf("Mismatch between expected setting %s and written setting %s", expected, given)
	}
}
