package spies

import (
	"github.com/continuouspipe/remote-environment-client/initialization"
)

type SpyInitState struct {
	Spy
	handle func() error
	next   func() initialization.InitState
	name   func() string
}

func NewSpyInitState() *SpyInitState {
	return &SpyInitState{}
}

func (s *SpyInitState) MockHandle(mocked func() error) {
	s.handle = mocked
}
func (s *SpyInitState) MockNext(mocked func() initialization.InitState) {
	s.next = mocked
}
func (s *SpyInitState) MockName(mocked func() string) {
	s.name = mocked
}

func (s *SpyInitState) Handle() error {
	args := make(Arguments)
	function := &Function{Name: "Handle", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.handle()
}
func (s *SpyInitState) Next() initialization.InitState {
	args := make(Arguments)
	function := &Function{Name: "Next", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.next()
}
func (s *SpyInitState) Name() string {
	args := make(Arguments)
	function := &Function{Name: "Name", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return s.name()
}
