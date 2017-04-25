//TODO: Refactor spies to use testify framework https://github.com/stretchr/testify
package spies

import "io"

//TODO: Update to use mock.Mock from testify framework https://github.com/stretchr/testify
type SpyInitStrategy struct {
	Spy
	complete  func(argsIn []string) error
	validate  func() error
	handle    func() error
	setWriter func(writer io.Writer)
}

func NewSpyInitStrategy() *SpyInitStrategy {
	return &SpyInitStrategy{}
}

func (s *SpyInitStrategy) MockComplete(mocked func(argsIn []string) error) {
	s.complete = mocked
}

func (s *SpyInitStrategy) MockValidate(mocked func() error) {
	s.validate = mocked
}

func (s *SpyInitStrategy) MockHandle(mocked func() error) {
	s.handle = mocked
}

func (s *SpyInitStrategy) Complete(argsIn []string) error {
	args := make(Arguments)
	args["argsIn"] = argsIn

	function := &Function{Name: "Complete", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.complete(argsIn)
}

func (s *SpyInitStrategy) Validate() error {
	args := make(Arguments)
	function := &Function{Name: "Validate", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.validate()
}

func (s *SpyInitStrategy) Handle() error {
	args := make(Arguments)
	function := &Function{Name: "Handle", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.handle()
}

func (s *SpyInitStrategy) SetWriter(writer io.Writer) {
	args := make(Arguments)
	args["writer"] = writer
	function := &Function{Name: "SetWriter", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
}
