package test

import (
	"testing"
	"reflect"
)

func AssertSame(t *testing.T, expected interface{}, actual interface{}) {
	if actual != expected {
		t.Errorf("Mismatch between expected setting %s and written setting %s", expected, actual)
	}
}

func AssertDeepEqual(t *testing.T, expected interface{}, actual interface{}) {
	//act := make([]string, len(actual))
	//exp := make([]string, len(expected))
	//copy(act, actual)
	//copy(exp, expected)
	//
	//sort.StringSlice(act).Sort()
	//sort.StringSlice(exp).Sort()

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected(sorted): %#v, Actual(sorted): %#v", expected, actual)
	}
}
