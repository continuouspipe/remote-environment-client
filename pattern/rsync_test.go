package pattern

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRsyncPathPattern_Match(t *testing.T) {
	scenarios := []struct {
		path        string
		patterns    []string
		description string
		toTransfer  bool
		err         error
	}{
		//This scenarios should match
		{
			"/Users/bob/dev/proj/path/to/file/a",
			[]string{"/Users/bob/dev/proj/path/to/file/a"},
			"file path exactly matches",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-1348",
			[]string{"+ file-1348"},
			"file path matches entry with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-9643",
			[]string{"proj"},
			"parent folder proj matches entry",
			true,
			nil,
		},

		{
			"/Users/bob/dev/proj/path/to/file/file-9643",
			[]string{"+ proj"},
			"parent folder proj matches entry with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"/Users/bob/dev/proj/path/to/file/*"},
			"star at the end of the partial matches",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"+ /Users/bob/dev/proj/path/to/file/*"},
			"star at the end of the partial matches with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"/Users/bob/dev/proj/path/to/*/file-2137"},
			"star on the parent folder of the partial matches",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"+ /Users/bob/dev/proj/path/to/*/file-2137"},
			"star on the parent folder of the partial matches with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"/Users/bob/dev/*/path/to/file/file-2137"},
			"star in the middle of the partial matches",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"+ /Users/bob/dev/*/path/to/file/file-2137"},
			"star in the middle of the partial matches with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"/Users/bob/dev/*"},
			"star in the middle of the partial pattern",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"+ /Users/bob/dev/*"},
			"star in the middle of the partial pattern with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-1348",
			[]string{"bob/dev/proj/path/to/file/file-1348"},
			"not anchored long pattern",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-1348",
			[]string{"+ bob/dev/proj/path/to/file/file-1348"},
			"not anchored long pattern with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-1348",
			[]string{"bob/dev/proj/path/*"},
			"not anchored short pattern with star",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-1348",
			[]string{"+ bob/dev/proj/path/*"},
			"not anchored short pattern with star with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-1348",
			[]string{"bob/dev/*/path/to/file/file-1348"},
			"not anchored short pattern with star in the middle",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-1348",
			[]string{"+ bob/dev/*/path/to/file/file-1348"},
			"not anchored short pattern with star in the middle with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"*"},
			"star pattern",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"+ *"},
			"star pattern with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"/Users/*/dev/*/path/*/file/*"},
			"multiple single star symbols in the pattern",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"+ /Users/*/dev/*/path/*/file/*"},
			"multiple single star symbols in the pattern with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"**"},
			"double star pattern",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"+ **"},
			"double star pattern with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"/Users/bob/dev/**/to/file/file-7144"},
			"double star pattern in the middle of the pattern",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"+ /Users/bob/dev/**/to/file/file-7144"},
			"double star pattern in the middle of the pattern with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"/Users/bob/**/**/to/file/file-7144"},
			"multiple double star pattern in the middle of the pattern",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"+ /Users/bob/**/**/to/file/file-7144"},
			"multiple double star pattern in the middle of the pattern with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"/Users/bob/**/file/file-7144"},
			"double star pattern in the middle of the pattern",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"+ /Users/bob/**/file/file-7144"},
			"double star pattern in the middle of the pattern with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"/Users/**/file-7144"},
			"double star pattern in the middle of the pattern",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"+ /Users/**/file-7144"},
			"double star pattern in the middle of the pattern with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"**/file/file-7144"},
			"double star pattern in the beginning of the pattern",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"+ **/file/file-7144"},
			"double star pattern in the beginning of the pattern with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"/Users/bob/**"},
			"double star pattern in the end of the pattern",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"+ /Users/bob/**"},
			"double star pattern in the end of the pattern with explicit inclusion",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"/Users/**/proj/path/**/file-7144"},
			"multiple double stars in pattern",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-7144",
			[]string{"+ /Users/**/proj/path/**/file-7144"},
			"multiple double stars in pattern with explicit inclusion",
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
		},
		{
			"/Users/bob/dev/proj/path/to/file/b",
			[]string{"- /Users/bob/dev/proj/path/to/file/a"},
			"file path does match and has explicit exclusion",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/b",
			[]string{"boo"},
			"simple pattern not anchored does not match",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/b",
			[]string{"PROJ"},
			"parent folder PROJ does not matches entry as is capital",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-9643",
			[]string{"/Users/bob/dev/proj/path/to/DIFFERENT/*"},
			"star at the end of the pattern does not match",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-9643",
			[]string{"- /Users/bob/dev/proj/path/to/file/*"},
			"star at the end of the pattern does matches and has explicit exclusion",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-9643",
			[]string{"/Users/bob/dev/DIFFERENT/path/file/*/file-9643"},
			"start in the file parent part of the pattern does not match",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-9643",
			[]string{"/Users/DIFFERENT/file/*/path/to/file/file-9643"},
			"star in the middel of the pattern does not match",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/file-2137",
			[]string{"/Users/DIFFERENT/dev/*"},
			"star in the middle of the partial pattern does not match",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{"/Users/bob/**/to/file/xyz"},
			"double star pattern in the middle of the pattern should not match",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{"**/dev/proj/path/to/file/xyz"},
			"double star pattern in the beginning of the pattern should not match",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{"- /Users/bob/dev/proj"},
			"parent folder is excluded",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{"- *"},
			"star pattern all folders excluded",
			false,
			nil,
		},
		{
			"/Users",
			[]string{"/Users/bob/dev/proj/path/to/file/a"},
			"parent path shoud not match nested file pattern",
			false,
			nil,
		},

		//Multiple patterns
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{
				"+ /Admin/something/else",
				"random-file",
				"random-folder/",
				"- /an-anchored/sub-folder/with/file/abc",
				"+ /*/bob/dev/proj/path/to/file/abc",
				"- /Users/*/dev/proj/path/to/file/abc",
				"+ /Users/bob/*/proj/path/to/file/abc",
				"- /Users/bob/dev/*/path/to/file/abc",
				"+ /Users/bob/dev/proj/*/to/file/abc",
				"- /Users/bob/dev/proj/path/*/file/abc",
				"+ /Users/bob/dev/proj/path/to/*/abc",
				"- /Users/bob/dev/proj/path/to/file/*",
			},
			"multiple patterns it should match '+ /*/bob/dev/proj/path/to/file/abc' and transfer",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{
				"+ /Admin/something/else",
				"random-file",
				"random-folder/",
				"- /an-anchored/sub-folder/with/file/abc",
				"- /Users/*/dev/proj/path/to/file/abc",
				"+ /Users/bob/*/proj/path/to/file/abc",
				"- /Users/bob/dev/*/path/to/file/abc",
				"+ /Users/bob/dev/proj/*/to/file/abc",
				"- /Users/bob/dev/proj/path/*/file/abc",
				"+ /Users/bob/dev/proj/path/to/*/abc",
				"- /Users/bob/dev/proj/path/to/file/*",
			},
			"multiple patterns it should match '- /Users/*/dev/proj/path/to/file/abc' and not transfer",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{
				"+ /Admin/something/else",
				"random-file",
				"random-folder/",
				"- /an-anchored/sub-folder/with/file/abc",
				"+ /Users/bob/*/proj/path/to/file/abc",
				"- /Users/bob/dev/*/path/to/file/abc",
				"+ /Users/bob/dev/proj/*/to/file/abc",
				"- /Users/bob/dev/proj/path/*/file/abc",
				"+ /Users/bob/dev/proj/path/to/*/abc",
				"- /Users/bob/dev/proj/path/to/file/*",
			},
			"multiple patterns it should match '+ /Users/bob/*/proj/path/to/file/abc' and transfer",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{
				"+ /Admin/something/else",
				"random-file",
				"- /an-anchored/sub-folder/with/file/abc",
				"- /Users/bob/dev/*/path/to/file/abc",
				"+ /Users/bob/dev/proj/*/to/file/abc",
				"- /Users/bob/dev/proj/path/*/file/abc",
				"+ /Users/bob/dev/proj/path/to/*/abc",
				"- /Users/bob/dev/proj/path/to/file/*",
			},
			"multiple patterns it should match '- /Users/bob/dev/*/path/to/file/abc' and not transfer",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{
				"+ /Admin/something/else",
				"random-file",
				"+ /Users/bob/dev/proj/*/to/file/abc",
				"- /Users/bob/dev/proj/path/*/file/abc",
				"+ /Users/bob/dev/proj/path/to/*/abc",
				"- /Users/bob/dev/proj/path/to/file/*",
			},
			"multiple patterns it should match '+ /Users/bob/dev/proj/*/to/file/abc' and transfer",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{
				"+ /Admin/something/else",
				"- /Users/bob/dev/proj/path/*/file/abc",
				"+ /Users/bob/dev/proj/path/to/*/abc",
				"- /Users/bob/dev/proj/path/to/file/*",
			},
			"multiple patterns it should match '- /Users/bob/dev/proj/path/*/file/abc' and not transfer",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{
				"+ /Users/bob/dev/proj/path/to/*/abc",
				"- /Users/bob/dev/proj/path/to/file/*",
			},
			"multiple patterns it should match '+ /Users/bob/dev/proj/path/to/*/abc' and transfer",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{
				"- /Users/bob/dev/proj/path/to/file/*",
				"- *",
			},
			"multiple patterns it should match '- /Users/bob/dev/proj/path/to/file/*' and not transfer due to parent rule '- *'",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{
				"+ /Users/bob/dev/proj/path/to/*/abc",
				"- *",
			},
			"multiple patterns it should match '+ /Users/bob/dev/proj/path/to/*/abc' and not transfer due to parent rule '- *'",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{
				"+ /Users/bob/dev/proj/path/to/file/abc",
				"- /Users/bob/dev/proj/path/to/file",
			},
			"multiple patterns it should match '+ /Users/bob/dev/proj/path/to/file/abc' and not transfer due to parent rule '- /Users/bob/dev/proj/path/to/file'",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{
				"+ /Users/bob/dev/proj/path/to/file/abc",
				"- /Users/bob/dev/proj",
			},
			"multiple patterns it should match '+ /Users/bob/dev/proj/path/to/file/abc' and not transfer due to parent rule '- /Users/bob/dev/proj'",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{
				"+ /Users/bob/dev/proj/path/to/file/abc",
				"- /Users",
			},
			"multiple patterns it should match '+ /Users/bob/dev/proj/path/to/file/abc' and not transfer due to parent rule '- /Users'",
			false,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{
				"+ /Users/bob/dev/proj/path/to/file/abc",
				"+ /Users",
			},
			"multiple patterns it should match '+ /Users/bob/dev/proj/path/to/file/abc' and transfer",
			true,
			nil,
		},
		{
			"/Users/bob/dev/proj/path/to/file/abc",
			[]string{
				"+ /Users/bob/dev/proj/path/to/file/abc",
				"+ *",
			},
			"multiple patterns it should match '+ /Users/bob/dev/proj/path/to/file/abc' and transfer",
			true,
			nil,
		},

		//Edge cases
		{
			"",
			[]string{"*"},
			"edge case: empty path and star pattern",
			false,
			errors.New("empty path given"),
		},
		{
			"",
			[]string{""},
			"edge case: empty path and empty pattern",
			false,
			errors.New("empty path given"),
		},
		{
			"abc",
			[]string{""},
			"edge case: random path and empty pattern",
			false,
			nil,
		},
	}

	subject := NewRsyncMatcherPath()

	for _, scenario := range scenarios {
		subject.AddPattern(scenario.patterns...)
		match, err := subject.IncludeToTransfer(scenario.path)
		assert.Equal(t, scenario.toTransfer, match, scenario.description)
		if scenario.err != nil {
			assert.Equal(t, scenario.err.Error(), err.Error(), scenario.description)
		}
	}
}
