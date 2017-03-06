package spies

//Spy for Push
type SpyPush struct {
	Spy
	push         func() (string, error)
	deleteRemote func(remoteName string, remoteBranch string) (string, error)
}

func NewSpyPush() *SpyPush {
	return &SpyPush{}
}

func (s *SpyPush) Push(localBranch string, remoteName string, remoteBranch string) (string, error) {
	args := make(Arguments)
	args["localBranch"] = localBranch
	args["remoteName"] = remoteName
	args["remoteBranch"] = remoteBranch

	function := &Function{Name: "Push", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.push()
}

func (s *SpyPush) MockPush(mocked func() (string, error)) {
	s.push = mocked
}

func (s *SpyPush) DeleteRemote(remoteName string, gitBranch string) (string, error) {
	args := make(Arguments)
	args["remoteName"] = remoteName
	args["gitBranch"] = gitBranch

	function := &Function{Name: "DeleteRemote", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.deleteRemote(remoteName, gitBranch)
}

func (s *SpyPush) MockDeleteRemote(mocked func(remoteName string, remoteBranch string) (string, error)) {
	s.deleteRemote = mocked
}
