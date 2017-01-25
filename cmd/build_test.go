package cmd

import (
	"testing"
	"github.com/continuouspipe/remote-environment-client/test"
)

func TestRemoteBranchNotPresent(t *testing.T) {
	//get mocked dependencies
	mocklsRemote := test.NewMockLsRemote()
	mocklsRemote.MockGetList(func(remoteName string, remoteBranch string) (string, error) {
		return "", nil
	})
	mockRevParse := test.NewMockRevParse()
	mockRevParse.MockGetLocalBranchName(func() (string, error) {
		return "feature-new", nil
	})
	spyCommit := test.NewSpyCommit()
	spyPush := test.NewSpyPush()

	//test subject called
	buildHandle := BuildHandle{}
	buildHandle.lsRemote = mocklsRemote
	buildHandle.revParse = mockRevParse
	buildHandle.commit = spyCommit
	buildHandle.push = spyPush
	buildHandle.remoteName = "origin"
	buildHandle.remoteBranch = "feature-my-remote"
	buildHandle.Handle()

	//expectations
	if spyCommit.CallsCountFor("Commit") != 0 {
		t.Error("Expected Commit not to be called")
	}

	if spyPush.CallsCountFor("Push") != 1 {
		t.Error("Expected Push to be called once")
	}
	firstCall := spyPush.FirstCallsFor("Push")
	if str, ok := firstCall.Arguments["localBranch"].(string); ok {
		test.AssertSame(t, "feature-new", str)
	} else {
		t.Fatalf("Expected local branch to be a string, given %T", firstCall.Arguments["localBranch"])
	}
	if str, ok := firstCall.Arguments["remoteName"].(string); ok {
		test.AssertSame(t, "origin", str)
	} else {
		t.Fatalf("Expected remote name to be a string, given %T", firstCall.Arguments["remoteName"])
	}

	if str, ok := firstCall.Arguments["remoteBranch"].(string); ok {
		test.AssertSame(t, "feature-my-remote", str)
	} else {
		t.Fatalf("Expected remote branch to be a string, given %T", firstCall.Arguments["remoteBranch"])
	}
}
