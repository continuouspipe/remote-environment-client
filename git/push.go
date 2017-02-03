// executes git diff commands
// e.g. git push --force "$(remote_name)" "$(local_branch):$(remote_branch)"
// e.g. git push "$(remote_name)" --delete "$(remote_branch)"
package git

import "github.com/continuouspipe/remote-environment-client/osapi"


type PushExecutor interface {
	Push(localBranch string, remoteName string, remoteBranch string) (string, error)
	DeleteRemote(remoteName string, remoteBranch string) (string, error)
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
	return osapi.CommandExec(getGitScmd(), args...)
}

func (g *push) DeleteRemote(remoteName string, remoteBranch string) (string, error) {
	args := []string{
		"push",
		remoteName,
		"--delete",
		remoteBranch,
	}
	return osapi.CommandExec(getGitScmd(), args...)
}
