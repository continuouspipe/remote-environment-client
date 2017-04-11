// Package pattern implements some of the rsync filter rules and include/exclude pattern rules
package pattern

import (
	"strings"
)

// PathPatternMatcher match a path against a list of patterns.
type PathPatternMatcher interface {
	SetPatterns(patterns []string)
	Match(path string) (matched bool, err error)
}

// RsyncMatcherPath allows to match a path using some of the rsync filter rules and include/exclude pattern rules
type RsyncMatcherPath struct {
	patterns []string
}

// NewRsyncMatcherPath ctor returns a pointer to RsyncMatcherPath
func NewRsyncMatcherPath() *RsyncMatcherPath {
	return &RsyncMatcherPath{}
}

// SetPatterns store the list of patterns
func (m *RsyncMatcherPath) SetPatterns(patterns []string) {
	m.patterns = patterns
}

// Match matches a path against a list of patterns.
//
// FILTER RULES (sub-set of the all supported rsync filter rules)
//
// - exclude, - specifies an exclude pattern.
// - include, + specifies an include pattern.
//
// INCLUDE/EXCLUDE PATTERN RULES  (sub-set of the all supported rsync include/exclude pattern rules)
// - if the pattern starts with a / then it is anchored to a particular spot in the hierarchy of files, otherwise it is matched against the end of the pathname.
// - a '*' matches any non-empty path component (it stops at slashes)
// - use '**' to match anything, including slashes
//
//
// Note  that, this is implemented as when rsync uses the --recursive (-r) option (which is implied by -a),
//
// Every subcomponent of every path is visited from the top down, so include/exclude patterns get applied recursively to each subcomponent's full name (e.g. to
// include "/foo/bar/baz" the subcomponents "/foo" and "/foo/bar" must not be excluded).
// The exclude patterns actually short-circuit the directory traversal stage when rsync finds the files to send.
// If a  pattern  excludes  a  particular parent directory, it can render a deeper include pattern ineffectual because rsync did not descend through that excluded section of the hierarchy.
// This is particularly important when using a trailing '*' rule.  For instance, this won't
// work:
//
// + /some/path/this-file-will-not-be-found
// + /file-is-included
// - *
func (m RsyncMatcherPath) Match(targetPath string) (matched bool, err error) {
	for _, pattern := range m.patterns {

		if strings.HasPrefix(targetPath, "/") && targetPath == pattern {
			return true, nil
		}

		targetPathElems := strings.Split(targetPath, "/")

		for _, targetPathElem := range targetPathElems {
			if targetPathElem == pattern {
				return true, nil
			}
		}
	}
	return false, nil
}