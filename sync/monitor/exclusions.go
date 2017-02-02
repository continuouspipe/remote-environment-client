//contains any logic that deals with what files should be excluded from the monitoring
package monitor

import (
	"bufio"
	"os"
	"regexp"
)

const CustomExclusionsFile = ".cp-remote-ignore"

type ExclusionProvider interface {
	LoadCustomExclusionsFromFile() error
	WriteDefaultExclusionsToFile() (bool, error)
	GetListExclusions() []string
	MatchExclusionList(target string) bool
}

type Exclusion struct {
	DefaultExclusions []string
	CustomExclusions  []string
}

func NewExclusion() *Exclusion {
	m := &Exclusion{}
	m.DefaultExclusions = []string{`/\.[^/]*$`,
		`\.idea`,
		`\.git`,
		`___jb_old___`,
		`___jb_tmp___`,
		`cp-remote-logs`,
		`.cp-remote-env-settings.yml`}
	return m
}

func (m *Exclusion) LoadCustomExclusionsFromFile() error {
	file, err := os.OpenFile(CustomExclusionsFile, os.O_RDWR|os.O_CREATE, 0664)
	defer file.Close()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		m.CustomExclusions = append(m.CustomExclusions, scanner.Text())
	}
	return nil
}

func (m *Exclusion) WriteDefaultExclusionsToFile() (bool, error) {
	file, err := os.OpenFile(CustomExclusionsFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	defer file.Close()
	if err != nil {
		return false, err
	}
	w := bufio.NewWriter(file)
	for _, line := range m.DefaultExclusions {
		_, err := w.WriteString(line)
		if err != nil {
			return false, err
		}
		w.WriteString("\n")
	}
	w.Flush()
	return true, nil
}

//returns the default exclusions if there aren't custom one loaded
func (m Exclusion) GetListExclusions() []string {
	if len(m.CustomExclusions) > 0 {
		return m.CustomExclusions
	}
	return m.DefaultExclusions
}

func (m Exclusion) MatchExclusionList(target string) bool {
	for _, elem := range m.GetListExclusions() {
		regex, err := regexp.Compile(elem)
		if err != nil {
			return false
		}
		if res := regex.MatchString(target); res == true {
			return true
		}
	}
	return false
}
