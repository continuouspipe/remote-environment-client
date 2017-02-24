package test

import (
	"reflect"
	"testing"
)

func AssertSame(t *testing.T, expected interface{}, actual interface{}) {
	if actual != expected {
		t.Errorf("Mismatch between expected setting: \n'%s'\n and written setting:\n'%s'", expected, actual)
	}
}

func AssertDeepEqual(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected: %#v, \nActual: %#v", expected, actual)
	}
}

func AssertError(t *testing.T, expected string, actual error) {
	if actual == nil {
		t.Errorf("Expected error %s", expected)
	}
	if expected != actual.Error() {
		t.Errorf("Mismatch between expected error: \n'%s'\nand actual error: \n'%s'", expected, actual)
	}
}

func AssertNotError(t *testing.T, actual error) {
	if actual != nil {
		t.Errorf("Unexpected error:\n'%s'", actual.Error())
	}
}
