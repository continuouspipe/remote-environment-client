// execute ls remote command
// e.g. git ls-remote --exit-code  --quiet . "origin/feature/cp-remote-testing"
package git

import "github.com/continuouspipe/remote-environment-client/osapi"

type LsRemoteExecutor interface {
	GetList(remoteName string, remoteBranch string) string
}

type lsRemote struct{}

func NewLsRemote() *lsRemote {
	return &lsRemote{}
}

func (g *lsRemote) GetList(remoteName string, remoteBranch string) (string, error) {
	args := []string{
		"ls-remote",
		"--exit-code",
		"--quiet",
		".",
		remoteName + "/" + remoteBranch,
	}

	return osapi.CommandExec("git", args...)
}
