package git

// libgit2/git2go has to be statically linked to libgit2
// brew install cmake
// go get -d github.com/libgit2/git2go
// git checkout next
// git submodule update --init # get libgit2
// make install
// see https://github.com/libgit2/git2go/blob/master/README.md for more info
import (
	"fmt"
	"os"
	
	lgit "github.com/libgit2/git2go"
)

type BranchRemover interface {
	Delete(remoteBranch string, remoteName string) (bool, error)
}

type GitBranchRemover struct{}

func NewGitBranchRemover() *GitBranchRemover {
	return &GitBranchRemover{}
}

func (g *GitBranchRemover) Delete(remoteBranch string, remoteName string) (bool, error) {
	_, err := lgit.OpenRepository("./")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return true, nil
}
