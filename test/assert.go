package test

import (
	"reflect"
	"runtime/debug"
	"testing"
)

//DEPRECATED use testify framework https://github.com/stretchr/testify
func AssertSame(t *testing.T, expected interface{}, actual interface{}) {
	if actual != expected {
		stack := debug.Stack()
		t.Errorf("Mismatch between expected setting: \n'%s'\n and written setting:\n'%s'\n%s", expected, actual, stack[:])
	}
}

//DEPRECATED use testify framework https://github.com/stretchr/testify
func AssertDeepEqual(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		stack := debug.Stack()
		t.Errorf("Expected: '%#v', \nActual: '%#v'\n%s", expected, actual, stack[:])
	}
}

//DEPRECATED use testify framework https://github.com/stretchr/testify
func AssertError(t *testing.T, expected string, actual error) {
	stack := debug.Stack()
	if actual == nil {
		t.Errorf("Expected error '%s'\n%s", expected, stack[:])
	}
	if expected != actual.Error() {
		t.Errorf("Mismatch between expected error: \n'%s'\nand actual error: \n'%s'\n%s", expected, actual, stack[:])
	}
}

//DEPRECATED use testify framework https://github.com/stretchr/testify
func AssertNotError(t *testing.T, actual error) {
	if actual != nil {
		stack := debug.Stack()
		t.Errorf("Unexpected error:\n'%s'\n%s", actual.Error(), stack[:])
	}
}
