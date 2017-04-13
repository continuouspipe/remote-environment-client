//Package monitor - exclusions contains any logic that deals with what files should be excluded from the monitoring
package monitor

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/pattern"
	"io"
	"os"
)

//CustomExclusionsFile is the default ignore file for rsync exclusions
const CustomExclusionsFile = ".cp-remote-ignore"

//ExclusionProvider allows to write an exclusion list to files and match a target against the exclusions
type ExclusionProvider interface {
	WriteDefaultExclusionsToFile() (bool, error)
	MatchExclusionList(target string) (bool, error)
}

//Exclusion contains elements that allow to store two list of path to exclude, a config.Ignore to load, save the
//exclusions and a pattern.PathPatternMatcher that allows to match a path to the exclusion list
type Exclusion struct {
	DefaultExclusions       []string
	FirstCreationExclusions []string
	ignore                  *config.Ignore
	rsyncMatcherPath        pattern.PathPatternMatcher
	writer                  io.Writer
}

//NewExclusion default constructor for Exclusion
func NewExclusion() *Exclusion {
	m := &Exclusion{}
	m.ignore = config.NewIgnore()
	m.ignore.File = CustomExclusionsFile
	m.rsyncMatcherPath = pattern.NewRsyncMatcherPath()
	m.writer = os.Stdout
	m.DefaultExclusions = []string{
		`.idea`,
		`.git`,
		`*___jb_old___`,
		`*___jb_tmp___`,
		`/cp-remote-logs**`,
		`/.cp-remote-settings.yml`,
		`/.cp-remote-env-settings.yml`}
	m.FirstCreationExclusions = []string{
		`.*`,
	}
	return m
}

// WriteDefaultExclusionsToFile check if the exclusion file exists
// if the exclusion file does not already exist in the system it will add the values contained on FirstCreationExclusions
// if the exclusion file already exists it simply add the missing DefaultExclusions using the config.Ignore struct
func (m *Exclusion) WriteDefaultExclusionsToFile() (bool, error) {
	exclusions := []string{}
	if _, err := os.Stat(m.ignore.File); os.IsNotExist(err) {
		exclusions = append(exclusions, m.FirstCreationExclusions...)
	}
	exclusions = append(exclusions, m.DefaultExclusions...)
	return m.ignore.AddToIgnore(exclusions...)
}

//MatchExclusionList loads the list of exclusions from the ignore file and
//feeds them to the NewRsyncMatcherPath
func (m Exclusion) MatchExclusionList(target string) (bool, error) {
	err := m.ignore.LoadFromIgnoreFile()
	if err != nil {
		return false, err
	}

	m.rsyncMatcherPath.AddPattern(m.ignore.List...)
	matchIncluded, msg, err := m.rsyncMatcherPath.HasMatchAndIsIncluded(target)
	if msg != "" {
		fmt.Fprintln(m.writer, msg)
	}
	if err != nil {
		cplogs.V(4).Infof("error when matching path to the exclusion list, details %s", err.Error())
		cplogs.Flush()
	}

	if matchIncluded {
		cplogs.V(5).Infof("the path %s is included in the transfer", target)
	} else {
		cplogs.V(5).Infof("the path %s is excluded from the transfer", target)
	}
	cplogs.Flush()
	return !matchIncluded, nil
}
