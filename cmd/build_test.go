package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/test/mocks"
	"github.com/continuouspipe/remote-environment-client/test/spies"
	"testing"
)

func TestRemoteBranchNotPresent(t *testing.T) {
	fmt.Println("Running TestRemoteBranchNotPresent")
	defer fmt.Println("TestRemoteBranchNotPresent Done")

	//get mocked dependencies
	mocklsRemote := mocks.NewMockLsRemote()
	mocklsRemote.MockGetList(func(remoteName string, remoteBranch string) (string, error) {
		return "", nil
	})
	mockRevParse := mocks.NewMockRevParse()
	mockRevParse.MockGetLocalBranchName(func() (string, error) {
		return "feature-new", nil
	})
	spyCommit := spies.NewSpyCommit()
	spyPush := spies.NewSpyPush()
	spyPush.MockPush(func() (string, error) {
		return "", nil
	})

	mockStdout := mocks.NewMockWriter()
	mockStdout.MockWrite(func(p []byte) (n int, err error) {
		return 100, nil
	})

	//test subject called
	buildHandle := BuildHandle{}
	buildHandle.lsRemote = mocklsRemote
	buildHandle.revParse = mockRevParse
	buildHandle.commit = spyCommit
	buildHandle.push = spyPush
	buildHandle.remoteName = "origin"
	buildHandle.remoteBranch = "feature-my-remote"
	buildHandle.Stdout = mockStdout
	buildHandle.Handle()

	//expectations
	spyCommit.ExpectsCallCount(t, "Commit", 0)
	spyPush.ExpectsCallCount(t, "Push", 1)
	spyPush.ExpectsFirstCallArgument(t, "Push", "localBranch", "feature-new")
	spyPush.ExpectsFirstCallArgument(t, "Push", "remoteName", "origin")
	spyPush.ExpectsFirstCallArgument(t, "Push", "remoteBranch", "feature-my-remote")
}
