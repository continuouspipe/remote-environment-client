package spies

//Spy for Commit
type SpyCommit struct {
	Spy
	commit func() (string, error)
}

func NewSpyCommit() *SpyCommit {
	return &SpyCommit{}
}

func (s *SpyCommit) Commit(message string) (string, error) {
	args := make(Arguments)
	args["message"] = message
	function := &Function{Name: "Commit", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.commit()
}

func (s *SpyCommit) SpyCommit(mocked func() (string, error)) {
	s.commit = mocked
}
