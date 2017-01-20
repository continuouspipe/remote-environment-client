// executes git diff command
// e.g. git push --force "$(remote_name)" "$(local_branch):$(remote_branch)"
package git

import (
	"github.com/continuouspipe/remote-environment-client/osapi"
)

type PushExecutor interface {
	Push(remoteName string, remoteBranch string) string
}

type push struct{}

func NewPush() *push {
    return &push{}
}

func (g *push) Push(localBranch string, remoteName string, remoteBranch string) (string, error) {
	args := []string{
		"push",
		"--force",
		remoteName,
		localBranch + ":" + remoteBranch,
	}

	return osapi.CommandExec("git", args...)
}
