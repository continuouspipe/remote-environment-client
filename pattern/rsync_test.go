package pattern

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRsyncPathPattern_Match(t *testing.T) {
	scenarios := []struct {
		path        string
		patterns    []string
		description string
		matched     bool
		err         error
	}{
		{
			"/Users/bob/dev/proj/path/to/file/a",
			[]string{"/Users/bob/dev/proj/path/to/file/a"},
			"file path exactly matches",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-1348",
			[]string{"file-1348"},
			"file path matches entry",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/b",
			[]string{"/Users/bob/dev/proj/path/to/file/a"},
			"file path does not match",
			false,
			nil,
		},
	}

	subject := NewRsyncMatcherPath()

	for _, scenario := range scenarios {
		subject.SetPatterns(scenario.patterns)
		match, err := subject.Match(scenario.path)
		assert.Equal(t, scenario.matched, match, scenario.description)
		assert.Equal(t, scenario.err, err, scenario.description)
	}
}
