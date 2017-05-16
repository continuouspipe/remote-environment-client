// executes git diff commands
// e.g. git push --force "$(remote_name)" "$(local_branch):$(remote_branch)"
// e.g. git push "$(remote_name)" --delete "$(remote_branch)"
package git

import (
	"net/http"

	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"github.com/pkg/errors"
)

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
	res, err := osapi.CommandExec(getGitScmd(), args...)
	if err != nil {
		return res, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "git error when force pushing to remote development branch").String())
	}
	return res, nil
}

func (g *push) DeleteRemote(remoteName string, remoteBranch string) (string, error) {
	args := []string{
		"push",
		remoteName,
		"--delete",
		remoteBranch,
	}
	res, err := osapi.CommandExec(getGitScmd(), args...)
	if err != nil {
		return res, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "git error when deleting the remote development branch").String())
	}
	return res, nil
}
