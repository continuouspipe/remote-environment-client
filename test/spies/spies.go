//TODO: Refactor spies to use testify framework https://github.com/stretchr/testify
package spies

import (
	"github.com/continuouspipe/remote-environment-client/test"
	"reflect"
	"testing"
)

//DEPRECATED use testify framework https://github.com/stretchr/testify
//A map that stores a list of function arguments [argumentName] => value (any type)
type Arguments map[string]interface{}

//DEPRECATED use testify framework https://github.com/stretchr/testify
//Function is a struct where you can set the name and add a slice Arguments ([]Argument) for each call
type Function struct {
	Name      string
	Arguments Arguments
}

//DEPRECATED use testify framework https://github.com/stretchr/testify
//Generic struct that can be embedded by any struct that wants to keep track to what function was called and with which args
type Spy struct {
	calledFunctions []Function
	commandExec     func() (string, error)
}

//DEPRECATED use testify framework https://github.com/stretchr/testify
func (spy *Spy) FirstCallsFor(functionName string) *Function {
	for _, call := range spy.calledFunctions {
		if call.Name == functionName {
			return &call
		}
	}
	return nil
}

//DEPRECATED use testify framework https://github.com/stretchr/testify
func (spy *Spy) NCallsFor(n int, functionName string) *Function {
	count := 0
	for _, call := range spy.calledFunctions {
		if call.Name == functionName {
			count = count + 1
			if count == n {
				return &call
			}
		}
	}
	return nil
}

//DEPRECATED use testify framework https://github.com/stretchr/testify
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

//DEPRECATED use testify framework https://github.com/stretchr/testify
func (spy *Spy) ExpectsCallCount(t *testing.T, functionName string, expectedCallCount int) {
	callsCount := spy.CallsCountFor(functionName)
	if callsCount != expectedCallCount {
		t.Errorf("Expected %s call count to be %d, actual call count %d", functionName, expectedCallCount, callsCount)
	}
}

//DEPRECATED use testify framework https://github.com/stretchr/testify
func (spy *Spy) ExpectsFirstCallArgument(t *testing.T, function string, argument string, expected interface{}) {
	firstCall := spy.FirstCallsFor(function)

	switch reflect.TypeOf(expected).Kind() {
	case reflect.Struct, reflect.Ptr:
		test.AssertDeepEqual(t, expected, firstCall.Arguments[argument])
	default:
		test.AssertSame(t, expected, firstCall.Arguments[argument])
	}
}

//DEPRECATED use testify framework https://github.com/stretchr/testify
func (spy *Spy) ExpectsCallNArgument(t *testing.T, function string, n int, argument string, expected interface{}) {
	nCall := spy.NCallsFor(n, function)

	switch reflect.TypeOf(expected).Kind() {
	case reflect.Struct, reflect.Ptr:
		test.AssertDeepEqual(t, expected, nCall.Arguments[argument])
	default:
		test.AssertSame(t, expected, nCall.Arguments[argument])
	}
}

//DEPRECATED use testify framework https://github.com/stretchr/testify
func (spy *Spy) ExpectsFirstCallArgumentStringSlice(t *testing.T, function string, argument string, expected []string) {
	firstCall := spy.FirstCallsFor(function)
	test.AssertDeepEqual(t, expected, firstCall.Arguments[argument])
}
