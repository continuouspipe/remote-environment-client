// executes git commit commands
// e.g. git commit --allow-empty -m "Add empty commit to force rebuild on continuous pipe"
package git

import "github.com/continuouspipe/remote-environment-client/osapi"

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

	return osapi.CommandExec("git", args...)
}
