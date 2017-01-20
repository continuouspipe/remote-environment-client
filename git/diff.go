// executes git diff command
// e.g. git diff --exit-code --quiet "feature/cp-remote-testing" "origin/feature/cp-remote-testing"
package git

import "github.com/continuouspipe/remote-environment-client/osapi"

type DiffExecutor interface {
	GetDiff(remoteName string, remoteBranch string) string
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
		remoteName + "/" + remoteBranch,
	}

	return osapi.CommandExec("git", args...)
}
