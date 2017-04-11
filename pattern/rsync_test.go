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
		}, {
			"/Users/bob/dev/proj/path/to/file/file-9643",
			[]string{"proj"},
			"parent folder proj matches entry",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"/Users/bob/dev/proj/path/to/file/*"},
			"star at the end of the partial matches",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"/Users/bob/dev/proj/path/to/*/file-2137"},
			"star on the parent folder of the partial matches ",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"/Users/bob/dev/*/path/to/file/file-2137"},
			"star in the middle of the partial matches",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"/Users/bob/dev/*"},
			"star in the middle of the partial pattern",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-1348",
			[]string{"bob/dev/proj/path/to/file/file-1348"},
			"not anchored long pattern",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-1348",
			[]string{"bob/dev/proj/path/*"},
			"not anchored short pattern with star",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-1348",
			[]string{"bob/dev/*/path/to/file/file-1348"},
			"not anchored short pattern with star in the middle",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"*"},
			"star pattern",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/b",
			[]string{"/Users/bob/dev/proj/path/to/file/a"},
			"file path does not match",
			false,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/b",
			[]string{"boo"},
			"simple pattern not anchored does not match",
			false,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/b",
			[]string{"PROJ"},
			"parent folder PROJ does not matches entry as is capital",
			false,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-9643",
			[]string{"/Users/bob/dev/proj/path/to/DIFFERENT/*"},
			"star at the end of the pattern does not match",
			false,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-9643",
			[]string{"/Users/bob/dev/DIFFERENT/path/file/*/file-9643"},
			"start in the file parent part of the pattern does not match",
			false,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-9643",
			[]string{"/Users/DIFFERENT/file/*/path/to/file/file-9643"},
			"star in the middel of the pattern does not match",
			false,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"/Users/DIFFERENT/dev/*"},
			"star in the middle of the partial pattern does not match",
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
