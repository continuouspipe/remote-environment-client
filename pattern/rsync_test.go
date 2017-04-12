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
		//This scenarios should match
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
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"/Users/*/dev/*/path/*/file/*"},
			"multiple single star symbols in the pattern",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"**"},
			"double star pattern",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"/Users/bob/dev/**/to/file/file-7144"},
			"double star pattern in the middle of the pattern",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"/Users/bob/**/**/to/file/file-7144"},
			"multiple double star pattern in the middle of the pattern",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"/Users/bob/**/file/file-7144"},
			"double star pattern in the middle of the pattern",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"/Users/**/file-7144"},
			"double star pattern in the middle of the pattern",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"**/file/file-7144"},
			"double star pattern in the beginning of the pattern",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"/Users/bob/**"},
			"double star pattern in the end of the pattern",
			true,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"/Users/**/proj/path/**/file-7144"},
			"multiple double stars in pattern",
			true,
			nil,
		},

		//This scenarios should not match
		{
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
		}, {
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{"/Users/bob/**/to/file/xyz"},
			"double star pattern in the middle of the pattern should not match",
			false,
			nil,
		}, {
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{"**/dev/proj/path/to/file/xyz"},
			"double star pattern in the beginning of the pattern should not match",
			false,
			nil,
		},

		//Multiple patterns
		//{
		//	"/Users/bob/dev/proj/path/to/file/abc",
		//	[]string{
		//		"+ /Admin/something/else",
		//		"random-file",
		//		"random-folder/",
		//		"- /an-anchored/sub-folder/with/file/abc",
		//		"+ */bob/dev/proj/path/to/file/abc",
		//		"- /Users/*/dev/proj/path/to/file/abc",
		//		"+ /Users/bob/*/proj/path/to/file/abc",
		//		"- /Users/bob/dev/*/path/to/file/abc",
		//		"+ /Users/bob/dev/proj/*/to/file/abc",
		//		"- /Users/bob/dev/proj/path/*/file/abc",
		//		"+ /Users/bob/dev/proj/path/to/*/abc",
		//		"- /Users/bob/dev/proj/path/to/file/*",
		//	},
		//	"double star pattern in the beginning of the pattern should not match",
		//	true,
		//	nil,
		//},
		//{
		//	"/Users/bob/dev/proj/path/to/file/abc",
		//	[]string{
		//		"+ /Admin/something/else",
		//		"random-file",
		//		"random-folder/",
		//		"- /an-anchored/sub-folder/with/file/abc",
		//		"- /Users/*/dev/proj/path/to/file/abc",
		//		"+ /Users/bob/*/proj/path/to/file/abc",
		//		"- /Users/bob/dev/*/path/to/file/abc",
		//		"+ /Users/bob/dev/proj/*/to/file/abc",
		//		"- /Users/bob/dev/proj/path/*/file/abc",
		//		"+ /Users/bob/dev/proj/path/to/*/abc",
		//		"- /Users/bob/dev/proj/path/to/file/*",
		//	},
		//	"double star pattern in the beginning of the pattern should not match",
		//	false,
		//	nil,
		//},

		//Edge cases
		{
			"",
			[]string{"*"},
			"edge case: empty path and star pattern",
			true,
			nil,
		}, {
			"",
			[]string{""},
			"edge case: empty path and empty pattern",
			true,
			nil,
		}, {
			"abc",
			[]string{""},
			"edge case: random path and empty pattern",
			false,
			nil,
		},
	}

	subject := NewRsyncMatcherPath()

	for _, scenario := range scenarios {
		subject.SetPatterns(scenario.patterns)
		match, err := subject.IncludeToTransfer(scenario.path)
		assert.Equal(t, scenario.matched, match, scenario.description)
		assert.Equal(t, scenario.err, err, scenario.description)
	}
}
