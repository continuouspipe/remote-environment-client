//TODO: Refactor mocks to use testify framework https://github.com/stretchr/testify
package mocks

//TODO: Update to use mock.Mock from testify framework https://github.com/stretchr/testify
//Mock for LsRemote
type MockLsRemote struct {
	getList func(remoteName string, remoteBranch string) (string, error)
}

func NewMockLsRemote() *MockLsRemote {
	return &MockLsRemote{}
}

func (m *MockLsRemote) MockGetList(mocked func(remoteName string, remoteBranch string) (string, error)) {
	m.getList = mocked
}

func (m *MockLsRemote) GetList(remoteName string, remoteBranch string) (string, error) {
	return m.getList(remoteName, remoteBranch)
}
