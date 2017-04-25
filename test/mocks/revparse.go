//TODO: Refactor mocks to use testify framework https://github.com/stretchr/testify
package mocks

//TODO: Update to use mock.Mock from testify framework https://github.com/stretchr/testify
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
