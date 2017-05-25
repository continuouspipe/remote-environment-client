package mocks

import (
	"github.com/continuouspipe/remote-environment-client/initialization"
	"github.com/stretchr/testify/mock"
)

//MockInitState is a Mock for InitState
type MockInitState struct {
	mock.Mock
}

//NewMockInitState Return an instance of MockInitState
func NewMockInitState() *MockInitState {
	return &MockInitState{}
}

//Handle records the arguments called and return the mocked arguments
func (h *MockInitState) Handle() (suggestion string, err error) {
	args := h.Called()
	return args.String(0), args.Error(1)
}

//Next records the arguments called and return the mocked arguments
func (h *MockInitState) Next() initialization.InitState {
	args := h.Called()
	return args.Get(0).(initialization.InitState)
}

//Name records the arguments called and return the mocked arguments
func (h *MockInitState) Name() string {
	args := h.Called()
	return args.String(0)
}
