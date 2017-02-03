// execute rev-parse commands
// e.g. git rev-parse --abbrev-ref HEAD
package git

import "github.com/continuouspipe/remote-environment-client/osapi"


type RevParseExecutor interface {
	GetLocalBranchName() (string, error)
}

type revParse struct{}

func NewRevParse() *revParse {
	return &revParse{}
}

func (g *revParse) GetLocalBranchName() (string, error) {
	args := []string{
		"rev-parse",
		"--abbrev-ref",
		"HEAD",
	}
	res, err := osapi.CommandExec(getGitScmd(), args...)
	if err != nil {
		return "", err
	}
	return res, nil
}
