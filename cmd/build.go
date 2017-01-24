package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/benchmark"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/git"
	"github.com/spf13/cobra"
)

func NewBuildCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "build",
		Aliases: []string{"bu"},
		Short:   "Create/Update the remote environment",
		Long: `The build command will push changes the branch you have checked out locally to your remote
environment branch. ContinuousPipe will then build the environment. You can use the
https://ui.continuouspipe.io/ to see when the environment has finished building and to
find its IP address.`,
		Run: func(cmd *cobra.Command, args []string) {
			settings := config.NewApplicationSettings()
			validator := config.NewMandatoryChecker()
			validateConfig(validator, settings)

			remoteName := settings.GetString(config.RemoteName)
			remoteBranch := settings.GetString(config.RemoteBranch)

			benchmrk := benchmark.NewCmdBenchmark()
			benchmrk.Start("build")

			handler := &BuildHandle{
				cmd,
				remoteName,
				remoteBranch,
			}
			handler.Handle(args)
			_, err := benchmrk.StopAndLog()
			checkErr(err)
		},
	}
}

type BuildHandle struct {
	Command      *cobra.Command
	remoteName   string
	remoteBranch string
}

func (h *BuildHandle) Handle(args []string) error {
	remoteExists := h.hasRemote()
	cplogs.V(5).Infof("remoteExists value is %s", remoteExists)

	if remoteExists == true {
		if localChanges := h.hasLocalChanges(); localChanges == false {
			h.commitAnEmptyChange()
		}
	}

	fmt.Println("Pushing to remote")
	h.pushToLocalBranch()
	fmt.Println("Continuous Pipe will now build your developer environment")
	fmt.Println("You can see when it is complete and find its IP address at https://ui.continuouspipe.io/")
	fmt.Println("Please wait until the build is complete to use any of this tool's other commands.")

	return nil
}

func (h *BuildHandle) pushToLocalBranch() {
	revparse := git.NewRevParse()
	push := git.NewPush()

	lbn, err := revparse.GetLocalBranchName()
	cplogs.V(5).Infof("local branch name value is %s", lbn)
	checkErr(err)

	push.Push(lbn, h.remoteName, h.remoteBranch)
}

func (h *BuildHandle) hasLocalChanges() bool {
	revparse := git.NewRevParse()
	lbn, err := revparse.GetLocalBranchName()
	cplogs.V(5).Infof("local branch name value is %s", lbn)
	checkErr(err)

	list := git.NewRevList()
	changes, err := list.GetLocalBranchAheadCount(lbn, h.remoteName, h.remoteBranch)
	cplogs.V(5).Infof("amount of changes found is %s", changes)
	checkErr(err)

	if changes > 0 {
		return true
	}
	return false
}

func (h *BuildHandle) hasRemote() bool {
	lsRemote := git.NewLsRemote()
	list, err := lsRemote.GetList(h.remoteName, h.remoteBranch)
	cplogs.V(5).Infof("list of remote branches that matches remote name and branch are %s", list)
	checkErr(err)
	if len(list) == 0 {
		return false
	}
	return true
}

func (h *BuildHandle) commitAnEmptyChange() {
	commit := git.NewCommit()
	fmt.Println("No changes so making an empty commit to force rebuild")
	_, err := commit.Commit("Add empty commit to force rebuild on continuous pipe")
	checkErr(err)
}
