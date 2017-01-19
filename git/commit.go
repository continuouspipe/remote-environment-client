package git

// libgit2/git2go has to be statically linked to libgit2
// brew install cmake
// go get -d github.com/libgit2/git2go
// git checkout next
// git submodule update --init # get libgit2
// make install
// see https://github.com/libgit2/git2go/blob/master/README.md for more info
import (
	"time"

	lgit "github.com/libgit2/git2go"
	"fmt"
)

type CommitTrigger interface {
	PushEmptyCommit(remoteBranch string, remoteName string) (bool, error)
}

type GitCommitTrigger struct{}

func NewGitCommitTrigger() *GitCommitTrigger {
	return &GitCommitTrigger{}
}

func (g *GitCommitTrigger) PushEmptyCommit(remoteBranch string, remoteName string) (bool, error) {

	repo, err := lgit.OpenRepository("./")
	if err != nil {
		return false, err
	}
	fmt.Printf("repo.Path() is %s", repo.Path())

	branch, err := repo.LookupBranch(remoteBranch, lgit.BranchRemote)
	if err != nil {
		return false, err
	}
	fmt.Printf("branch.Name() is %s", branch.Name())

	idx, err := repo.Index()
	if err != nil {
		return false, err
	}
	fmt.Printf("idx is %d", idx)

	//err = idx.AddByPath("some content")
	//if err != nil {
	//	panic(err)
	//}

	treeId, err := idx.WriteTree()
	if err != nil {
		return false, err
	}
	fmt.Printf("treeId is %d", treeId)

	err = idx.Write()
	if err != nil {
		return false, err
	}

	tree, err := repo.LookupTree(treeId)
	if err != nil {
		return false, err
	}

	commitTarget, err := repo.LookupCommit(branch.Target())
	if err != nil {
		return false, err
	}
	fmt.Printf("commitTarget.Id() is %d", commitTarget.Id())
	fmt.Printf("commitTarget.Message() is %s", commitTarget.Message())
	fmt.Printf("commitTarget.Summary() is %s", commitTarget.Summary())

	signature := &lgit.Signature{
		Name:  "Continuous Pipe",
		Email: "helpdesk@continuouspipe.com",
		When:  time.Now(),
	}

	oid, err := repo.CreateCommit("refs/heads/"+remoteBranch, signature, signature, "Add empty commit to force rebuild on continuous pipe", tree, commitTarget)
	if err != nil {
		return false, err
	}
	fmt.Printf("oid is %d", oid)

	return true, nil
}
