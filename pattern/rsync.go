// Package pattern implements some of the rsync filter rules and include/exclude pattern rules
package pattern

import (
	"errors"
	"fmt"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"path"
	"strings"
)

const (
	filterRuleInclude = iota
	filterRuleExclude
)

type pathPatternItem struct {
	prefix     rune
	pattern    string
	rawPattern string
}

// PathPatternMatcher match a path against a list of patterns.
type PathPatternMatcher interface {
	AddPattern(pattern ...string)
	HasMatchAndIsIncluded(path string) (include bool, message string, err error)
}

// RsyncMatcherPath allows to match a path using some of the rsync filter rules and include/exclude pattern rules
type RsyncMatcherPath struct {
	patternItems []pathPatternItem
}

// NewRsyncMatcherPath ctor returns a pointer to RsyncMatcherPath
func NewRsyncMatcherPath() *RsyncMatcherPath {
	return &RsyncMatcherPath{}
}

// AddPattern convert a pattern string into a pathPatternItem struct and stores it in an array
func (m *RsyncMatcherPath) AddPattern(pattern ...string) {
	var patternItems []pathPatternItem
	for _, p := range pattern {

		if len(p) == 0 {
			continue
		}

		patternItem := pathPatternItem{}
		patternItem.rawPattern = p
		patternItem.pattern = p
		//default prefix exclude
		patternItem.prefix = filterRuleExclude

		if p[0] == '+' && p[1] == ' ' {
			patternItem.prefix = filterRuleInclude
			patternItem.pattern = p[2:]
		}
		if p[0] == '-' && p[1] == ' ' {
			patternItem.pattern = p[2:]
		}
		patternItems = append(patternItems, patternItem)
	}
	m.patternItems = patternItems
}

// HasMatchAndIsIncluded determins if a
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
func (m *RsyncMatcherPath) HasMatchAndIsIncluded(path string) (include bool, details string, err error) {
	if len(path) == 0 {
		err := errors.New("empty path given")
		msg := fmt.Sprintf("error: %s", err.Error())
		return false, msg, err
	}
	if len(m.patternItems) == 0 {
		return true, "", nil
	}

	inc, found := m.filteredMatch(path)
	if inc == false && found != nil {
		msg := fmt.Sprintf("Not transferring %s because of pattern %s", path, found.rawPattern)
		return false, msg, nil
	}

	//before saying that we can safely include a path we need to check that none of is parent
	//has been excluded
	if inc == true && found != nil {
		//check if any of the parent folder is excluded
		parts := strings.Split(path, "/")
		current := ""
		for i := 1; i < len(parts)-1; i++ {
			current = current + "/" + parts[i]

			inc, found := m.filteredMatch(current)

			if inc == false && found != nil {
				msg := fmt.Sprintf("Not transferring %s because of pattern %s", path, found.rawPattern)
				return false, msg, nil
			}
		}
	}
	return true, "", nil
}

func (m RsyncMatcherPath) filteredMatch(path string) (include bool, found *pathPatternItem) {
	matchedPatterns := m.match(path)
	if len(matchedPatterns) > 0 {
		switch matchedPatterns[0].prefix {
		case filterRuleExclude:
			cplogs.V(5).Infof("not transferring %s because of the first pattern found: %s. List of all matches: %#v", path, matchedPatterns[0].rawPattern, matchedPatterns)
			cplogs.Flush()
			return false, &matchedPatterns[0]
		case filterRuleInclude:
			return true, &matchedPatterns[0]
		}
	}
	return true, nil
}

// Match matches a path against a list of patterns.
func (m RsyncMatcherPath) match(targetPath string) (matchedPatterns []pathPatternItem) {
	for _, patternItem := range m.patternItems {
		if strings.HasPrefix(patternItem.pattern, "/") {
			if res := m.matchedAnchoredPattern(targetPath, patternItem.pattern); res == true {
				matchedPatterns = append(matchedPatterns, patternItem)
			}
		} else if res := m.matchedRelativePattern(targetPath, patternItem.pattern); res == true {
			matchedPatterns = append(matchedPatterns, patternItem)
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
	//remove first empty element
	if len(targetElems) > 0 && targetElems[0] == "" {
		targetElems = targetElems[1:]
	}

	for _, targetPathElem := range targetElems {
		if targetPathElem == pattern {
			return true
		}
	}

	patternElems := strings.Split(pattern, "/")

	//remove first empty element
	if len(patternElems) > 0 && patternElems[0] == "" {
		patternElems = patternElems[1:]
	}

	//when there is only 1 pattern element and is not anchored
	//we need to match it against all patternElems and only return a failure if all of them don't match
	if len(patternElems) == 1 {
		return m.singlePartMatches(targetElems, patternElems[0])
	}

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

//singlePartMatches matches a single pattern against a list of targets
func (m RsyncMatcherPath) singlePartMatches(target []string, pattern string) (matches bool) {
	if pattern == "*" || pattern == "**" {
		return true
	}

	if strings.ContainsRune(pattern, '*') {
		return matchAny(pattern, target)
	}
	return false
}

//sequentialPartMatches iterates until we either reach the end of the pattern or we reach the end of the target string
func (m RsyncMatcherPath) sequentialPartMatches(target string, pattern string, offset int) (matches bool) {
	targetElems := strings.Split(target, "/")
	patternElems := strings.Split(pattern, "/")

	//remove first empty element
	if len(targetElems) > 0 && targetElems[0] == "" {
		targetElems = targetElems[1:]
	}
	if len(patternElems) > 0 && patternElems[0] == "" {
		patternElems = patternElems[1:]
	}

	patternKey := 0
	targetKey := 0 + offset

	//by default we return that the target matches the pattern
	matches = true

	for patternKey < len(patternElems) && targetKey < len(targetElems) {
		patternElem := patternElems[patternKey]
		targetElem := targetElems[targetKey]

		//check if the patternElem doesn't match the targetElem
		//e.g. target: /user/a, targetElem: a,
		//     pattern: /user/b, patternEle: b
		if patternElem != targetElem && !strings.ContainsRune(patternElem, '*') && patternElem != "**" {
			matches = false
			break
		}

		//if patternElem is ** we need to skip to the next targetElement
		//e.g. target:              /user/a/b/c/d/d/e/f
		//     targetKey:    -------------^
		//     pattern:             /user/**/d/e/e/f
		//     patternKey:   -------------^
		// targetKey will point to b, c until 'd' is reached which is the next element after **
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
			//
			// when patternKey reaches 2, the value would be ** and the next valid pattern key is 'h'
			if patternElems[nextValidPatternElemKey] == targetElem {
				patternKey = nextValidPatternElemKey + 1
			}

		} else {
			patternKey++
		}

		targetKey++
	}

	//we reached the end of the target string, and there are still other pattern elements to match
	//e.g. target: /user/a/b, pattern: /user/a/b/**
	//or   target: /user/a/b, pattern: /user/a/b/c/d/e/*/g
	if patternKey < len(patternElems) && targetKey == len(targetElems) {
		patternElem := patternElems[patternKey]

		//if the current pattern element is a double star symbol
		//break as it would match anything after
		if patternElem == "**" {
			//there are other pattern after '**'
			if patternKey+1 == len(patternElems) {
				//no other patterns, return true
				matches = true
			} else {
				//there are other patterns
				matches = false
			}
		} else {
			//else it means that we had other parts of the pattern that had to be matched
			//but the target string doesn't contain enough parts
			matches = false
		}

	}

	return
}

func matchAny(pattern string, items []string) (matched bool) {
	for _, item := range items {
		matches, _ := path.Match(pattern, item)

		if matches == false {
			continue
		}
		return true
	}
	return false
}
