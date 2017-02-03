// execute rev-list commands
// e.g. git rev-list --count origin/feature/cp-remote-testing...feature/cp-remote-testing
package git

import (
	"fmt"
	"strconv"

	"github.com/continuouspipe/remote-environment-client/osapi"
)

type RevListExecutor interface {
	GetLocalBranchAheadCount(localBranch string, remoteName string, remoteBranch string) (int, error)
}

type revList struct{}

func NewRevList() *revList {
	return &revList{}
}
func (g *revList) GetLocalBranchAheadCount(localBranch string, remoteName string, remoteBranch string) (int, error) {
	args := []string{
		"rev-list",
		"--count",
		fmt.Sprintf("%s/%s...%s", remoteName, remoteBranch, localBranch),
	}
	scount, err := osapi.CommandExec(getGitScmd(), args...)
	if err != nil {
		return 0, err
	}
	count, err := strconv.Atoi(scount)
	return count, err
}
