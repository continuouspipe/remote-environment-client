package mocks

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
