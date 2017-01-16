package test

import "testing"

func AssertSame(t *testing.T, expected string, given string) {
	if given != expected {
		t.Errorf("Mismatch between expected setting %s and written setting %s", expected, given)
	}
}
