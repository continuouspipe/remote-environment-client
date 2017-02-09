package mocks

//Mock for RevParse
type MockRevParse struct {
	getLocalBranchName func() (string, error)
}

func NewMockRevParse() *MockRevParse {
	return &MockRevParse{}
}

func (m *MockRevParse) MockGetLocalBranchName(mocked func() (string, error)) {
	m.getLocalBranchName = mocked
}

func (m *MockRevParse) GetLocalBranchName() (string, error) {
	return m.getLocalBranchName()
}
