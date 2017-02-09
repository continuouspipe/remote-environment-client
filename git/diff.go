// executes git diff commands
// e.g. git diff --exit-code --quiet "feature/cp-remote-testing" "origin/feature/cp-remote-testing"
package git

import (
	"github.com/continuouspipe/remote-environment-client/osapi"
	"path/filepath"
)

type DiffExecutor interface {
	GetDiff(remoteName string, remoteBranch string) (string, error)
}

type diff struct{}

func NewDiff() *diff {
	return &diff{}
}

func (g *diff) GetDiff(remoteName string, remoteBranch string) (string, error) {
	args := []string{
		"diff",
		"--exit-code",
		"--quiet",
		remoteBranch,
		remoteName + string(filepath.Separator) + remoteBranch,
	}
	return osapi.CommandExec(getGitScmd(), args...)
}
