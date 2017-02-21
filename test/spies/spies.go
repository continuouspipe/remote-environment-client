package spies

import (
	"github.com/continuouspipe/remote-environment-client/test"
	"reflect"
	"testing"
)

//A map that stores a list of function arguments [argumentName] => value (any type)
type Arguments map[string]interface{}

//Function is a struct where you can set the name and add a slice Arguments ([]Argument) for each call
type Function struct {
	Name      string
	Arguments Arguments
}

//Generic struct that can be embedded by any struct that wants to keep track to what function was called and with which args
type Spy struct {
	calledFunctions []Function
	commandExec     func() (string, error)
}

//Returns the Function call element for the given functionName, this is useful when a type has received multiple functions
func (spy *Spy) FirstCallsFor(functionName string) *Function {
	for _, call := range spy.calledFunctions {
		if call.Name == functionName {
			return &call
		}
	}
	return nil
}

//returns how many times the given function has been called
func (spy *Spy) CallsCountFor(functionName string) int {
	count := 0
	for _, call := range spy.calledFunctions {
		if call.Name != functionName {
			continue
		}
		count++
	}
	return count
}

func (spy *Spy) ExpectsCallCount(t *testing.T, functionName string, expectedCallCount int) {
	if spy.CallsCountFor(functionName) != expectedCallCount {
		t.Errorf("Expected %s call count to be %d", functionName, expectedCallCount)
	}
}

func (spy *Spy) ExpectsFirstCallArgument(t *testing.T, function string, argument string, expected interface{}) {
	firstCall := spy.FirstCallsFor(function)

	switch reflect.TypeOf(expected).Kind() {
	case reflect.Struct, reflect.Ptr:
		test.AssertDeepEqual(t, expected, firstCall.Arguments[argument])
	default:
		test.AssertSame(t, expected, firstCall.Arguments[argument])
	}
}

func (spy *Spy) ExpectsFirstCallArgumentStringSlice(t *testing.T, function string, argument string, expected []string) {
	firstCall := spy.FirstCallsFor(function)
	test.AssertDeepEqual(t, expected, firstCall.Arguments[argument])
}
