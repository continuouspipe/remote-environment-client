// Package pattern implements some of the rsync filter rules and include/exclude pattern rules
package pattern

import (
	"strings"
)

// PathPatternMatcher match a path against a list of patterns.
type PathPatternMatcher interface {
	SetPatterns(patterns []string)
	IncludeToTransfer(path string) (include bool)
	match(path string) (pattern string)
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

// IncludeToTransfer determins if a
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
func (m *RsyncMatcherPath) IncludeToTransfer(path string) (include bool, err error) {
	matchedPatterns := m.match(path)
	if len(matchedPatterns) > 0 {
		return true, nil
	}
	return false, nil
}

// Match matches a path against a list of patterns.
func (m RsyncMatcherPath) match(targetPath string) (matchedPatterns []string) {
	for _, p := range m.patterns {
		if strings.HasPrefix(p, "/") {
			if res := m.matchedAnchoredPattern(targetPath, p); res == true {
				matchedPatterns = append(matchedPatterns, p)
			}
		}
		if res := m.matchedRelativePattern(targetPath, p); res == true {
			matchedPatterns = append(matchedPatterns, p)
		}
	}
	return matchedPatterns
}

func (m RsyncMatcherPath) matchedAnchoredPattern(targetPath string, pattern string) (matched bool) {
	if targetPath == pattern {
		return true
	}
	if matches := m.sequentialPartMatches(targetPath, pattern, 0); matches {
		return true
	}
	return false
}

func (m RsyncMatcherPath) matchedRelativePattern(targetPath string, pattern string) (matched bool) {

	//Iterate recursively through all the target path elements and return true if one of them matches the given pattern
	targetElems := strings.Split(targetPath, "/")
	for _, targetPathElem := range targetElems {
		if targetPathElem == pattern {
			return true
		}
	}

	patternElems := strings.Split(pattern, "/")

	//Find the first matching pattern element and store is key in the offset
	offset := 0
	for key, targetElem := range targetElems {
		if targetElem == patternElems[0] {
			offset = key
		}
	}

	//Find how many parts match the given pattern taking in consideration the calcualted offset
	if matches := m.sequentialPartMatches(targetPath, pattern, offset); matches {
		return true
	}

	return false
}

func (m RsyncMatcherPath) sequentialPartMatches(target string, pattern string, offset int) (matches bool) {
	targetElems := strings.Split(target, "/")
	patternElems := strings.Split(pattern, "/")

	patternKey := 0
	targetKey := 0 + offset

	for patternKey < len(patternElems) && targetKey < len(targetElems) {
		patternElem := patternElems[patternKey]
		targetElem := targetElems[targetKey]

		if patternElem != targetElem && patternElem != "*" && patternElem != "**" {
			return false
		}

		if patternElem == "**" {
			//iterate to the next pattern part only if the next patternElem is not "*" or "**" and it matches the targetElem
			nextValidPatternElemKey := 0

			for i := patternKey; i < len(patternElems); i++ {
				if patternElems[i] != "*" && patternElems[i] != "**" {
					nextValidPatternElemKey = i
					break
				}
			}

			//if the target path element is the same as the next valid pattern element increment
			//ex.:
			// given path /a/b/c/d/e/f/g/h/i
			// and patternElems /a/b/**/**/h/i
			// when patternKey reaches 2, the value would be ** and the next valid pattern key is 'h'
			if patternElems[nextValidPatternElemKey] == targetElem {
				patternKey = nextValidPatternElemKey + 1
			}

		} else {
			patternKey++
		}

		targetKey++
	}

	return true
}
