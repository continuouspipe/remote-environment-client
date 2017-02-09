// executes git commit commands
// e.g. git commit --allow-empty -m "Add empty commit to force rebuild on continuous pipe"
package git

import (
	"github.com/continuouspipe/remote-environment-client/osapi"
	"os"
)

type CommitExecutor interface {
	Commit(message string) (string, error)
}

type commit struct{}

func NewCommit() *commit {
	return &commit{}
}

func (g *commit) Commit(message string) (string, error) {
	args := []string{
		"commit",
		"--allow-empty",
		"-m",
		message,
	}
	return osapi.CommandExec(getGitScmd(), args...)
}
func getGitScmd() osapi.SCommand {
	scmd := osapi.SCommand{}
	scmd.Name = "git"
	scmd.Stdin = os.Stdin
	scmd.Stdout = os.Stdout
	scmd.Stderr = os.Stderr
	return scmd
}
