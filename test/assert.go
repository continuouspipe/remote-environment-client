package test

import (
	"reflect"
	"testing"
)

func AssertSame(t *testing.T, expected interface{}, actual interface{}) {
	if actual != expected {
		t.Errorf("Mismatch between expected setting %s and written setting %s", expected, actual)
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
		t.Errorf("Mismatch between expected error %s and actual error %s", expected, actual)
	}
}

func AssertNotError(t *testing.T, actual error) {
	if actual != nil {
		t.Errorf("Unexpected error %s", actual.Error())
	}
}
