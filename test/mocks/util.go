package mocks

//Mock for QuestionPrompt
type MockQuestionPrompt struct{}

func NewMockQuestionPrompt() *MockQuestionPrompt {
	return &MockQuestionPrompt{}
}

func (qp MockQuestionPrompt) ReadString(q string) string {
	questions := [11]struct {
		question, answer string
	}{
		{"What is your Continuous Pipe project key?", "my-project"},
		{"What is the name of the Git branch you are using for your remote environment?", "feature/MYPROJ-312-initial-commit"},
		{"What is your github remote name? (defaults to: origin)", ""},
		{"What is the default container for the watch, bash, fetch and resync commands?", "web"},
		{"What is the IP of the cluster?", "127.0.0.1"},
		{"What is the cluster username?", "root"},
		{"What is the cluster password?", "2e9fik2s9-fds903"},
		{"If you want to use AnyBar, please provide a port number e.g 1738 ?", "6542"},
		{"What is your keen.io write key? (Optional, only needed if you want to record usage stats)", "sk29dj22d882"},
		{"What is your keen.io project id? (Optional, only needed if you want to record usage stats)", "cc3d902idi01"},
		{"What is your keen.io event collection?  (Optional, only needed if you want to record usage stats)", "event-collection"},
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

func (qp MockQuestionPrompt) RepeatUntilValid(question string, isValid func(string) (bool, error)) string {
	return qp.ReadString(question)
}
