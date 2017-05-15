// execute ls remote commands
// e.g. git ls-remote --exit-code  --quiet . "origin/feature/cp-remote-testing"
package git

import (
	"net/http"
	"path/filepath"

	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"github.com/pkg/errors"
)

type LsRemoteExecutor interface {
	GetList(remoteName string, remoteBranch string) (string, error)
}

type lsRemote struct{}

func NewLsRemote() *lsRemote {
	return &lsRemote{}
}

func (g *lsRemote) GetList(remoteName string, remoteBranch string) (string, error) {
	args := []string{
		"ls-remote",
		"--quiet",
		".",
		remoteName + string(filepath.Separator) + remoteBranch,
	}
	res, err := osapi.CommandExec(getGitScmd(), args...)
	if err != nil {
		return res, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "error when getting the list of remote branches using git ls-remote command").String())
	}
	return res, nil
}
