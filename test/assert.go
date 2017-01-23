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
		t.Errorf("Expected(un-sorted): %#v, \nActual(un-sorted): %#v", expected, actual)
	}
}
