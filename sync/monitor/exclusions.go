//contains any logic that deals with what files should be excluded from the monitoring
package monitor

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"os"
	"regexp"
	"strings"
)

const CustomExclusionsFile = ".cp-remote-ignore"

type ExclusionProvider interface {
	WriteDefaultExclusionsToFile() (bool, error)
	MatchExclusionList(target string) (bool, error)
}

type Exclusion struct {
	DefaultExclusions       []string
	FirstCreationExclusions []string
	ignore                  *config.Ignore
}

func NewExclusion() *Exclusion {
	m := &Exclusion{}
	m.ignore = config.NewIgnore()
	m.ignore.File = CustomExclusionsFile
	m.DefaultExclusions = []string{
		`.idea`,
		`.git`,
		`___jb_old___`,
		`___jb_tmp___`,
		`cp-remote-logs`,
		`.cp-remote-settings.yml`,
		`.cp-remote-env-settings.yml`}
	m.FirstCreationExclusions = []string{
		`.*`,
	}
	return m
}

func (m *Exclusion) WriteDefaultExclusionsToFile() (bool, error) {
	exclusions := []string{}
	if _, err := os.Stat(m.ignore.File); os.IsNotExist(err) {
		exclusions = append(exclusions, m.FirstCreationExclusions...)
	}
	exclusions = append(exclusions, m.DefaultExclusions...)
	return m.ignore.AddToIgnore(exclusions...)
}

func (m Exclusion) MatchExclusionList(target string) (bool, error) {
	err := m.ignore.LoadFromIgnoreFile()
	if err != nil {
		return false, err
	}

	for _, elem := range m.ignore.List {
		//if is an exact match return true
		if elem == target {
			return true, nil
		}

		//otherwise escape the "." and try to match it as a regex
		escapedElem := strings.Replace(elem, ".", `\.`, -1)
		regex, err := regexp.Compile(escapedElem)
		if err != nil {
			return false, nil
		}
		if res := regex.MatchString(target); res == true {
			return true, nil
		}
	}
	return false, nil
}
