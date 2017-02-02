package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/benchmark"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/git"
	"github.com/spf13/cobra"
	"strings"
	"io"
	"os"
)

func NewBuildCmd() *cobra.Command {
	settings := config.NewApplicationSettings()
	handler := &BuildHandle{}
	handler.commit = git.NewCommit()
	handler.lsRemote = git.NewLsRemote()
	handler.push = git.NewPush()
	handler.revList = git.NewRevList()
	handler.revParse = git.NewRevParse()
	handler.Stdout = os.Stdout

	command := &cobra.Command{
		Use:     "build",
		Aliases: []string{"bu"},
		Short:   "Create/Update the remote environment",
		Long: `The build command will push changes the branch you have checked out locally to your remote
environment branch. ContinuousPipe will then build the environment. You can use the
https://ui.continuouspipe.io/ to see when the environment has finished building and to
find its IP address.`,
		Run: func(cmd *cobra.Command, args []string) {
			validator := config.NewMandatoryChecker()
			validateConfig(validator, settings)

			benchmrk := benchmark.NewCmdBenchmark()
			benchmrk.Start("build")

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle())
			_, err := benchmrk.StopAndLog()
			checkErr(err)
		},
	}
	return command
}

type BuildHandle struct {
	Command      *cobra.Command
	commit       git.CommitExecutor
	lsRemote     git.LsRemoteExecutor
	push         git.PushExecutor
	revList      git.RevListExecutor
	revParse     git.RevParseExecutor
	remoteName   string
	remoteBranch string
	Stdout       io.Writer
}

// Complete verifies command line arguments and loads data from the command environment
func (h *BuildHandle) Complete(cmd *cobra.Command, argsIn []string, settingsReader config.Reader) error {
	h.Command = cmd
	h.remoteName = settingsReader.GetString(config.RemoteName)
	h.remoteBranch = settingsReader.GetString(config.RemoteBranch)
	return nil
}

// Validate checks that the provided build options are specified.
func (h *BuildHandle) Validate() error {
	if len(strings.Trim(h.remoteName, " ")) == 0 {
		return fmt.Errorf("the remote name specified is invalid")
	}
	if len(strings.Trim(h.remoteBranch, " ")) == 0 {
		return fmt.Errorf("the remote branch specified is invalid")
	}
	return nil
}

// Handle triggers a build on CP doing an empty commit on the given branch
// The empty commit will create a remote branch if it does not exist yet
// If there is a local commit ready to be pushed, it pushes those changes
func (h *BuildHandle) Handle() error {
	remoteExists, err := h.hasRemote()
	if err != nil {
		return err
	}
	cplogs.V(5).Infof("remoteExists value is %b", remoteExists)

	if remoteExists == true {
		localChanges, err := h.hasLocalChanges()
		if err != nil {
			return err
		}
		if localChanges == false {
			h.commitAnEmptyChange()
		}
	}

	fmt.Fprintln(h.Stdout, "Pushing to remote")
	h.pushToLocalBranch()
	fmt.Fprintln(h.Stdout, "Continuous Pipe will now build your developer environment")
	fmt.Fprintln(h.Stdout, "You can see when it is complete and find its IP address at https://ui.continuouspipe.io/")
	fmt.Fprintln(h.Stdout, "Please wait until the build is complete to use any of this tool's other commands.")

	return nil
}

func (h *BuildHandle) pushToLocalBranch() error {
	lbn, err := h.revParse.GetLocalBranchName()
	cplogs.V(5).Infof("local branch name value is %s", lbn)
	if err != nil {
		return err
	}

	h.push.Push(lbn, h.remoteName, h.remoteBranch)
	return nil
}

func (h *BuildHandle) hasLocalChanges() (bool, error) {
	lbn, err := h.revParse.GetLocalBranchName()
	cplogs.V(5).Infof("local branch name value is %s", lbn)
	if err != nil {
		return false, err
	}

	changes, err := h.revList.GetLocalBranchAheadCount(lbn, h.remoteName, h.remoteBranch)
	cplogs.V(5).Infof("amount of changes found is %s", changes)
	if err != nil {
		return false, err
	}

	if changes > 0 {
		return true, nil
	}

	return false, nil
}

func (h *BuildHandle) hasRemote() (bool, error) {
	list, err := h.lsRemote.GetList(h.remoteName, h.remoteBranch)
	cplogs.V(5).Infof("list of remote branches that matches remote name and branch are %s", list)
	if err != nil {
		return false, err
	}
	if len(list) == 0 {
		return false, err
	}
	return true, err
}

func (h *BuildHandle) commitAnEmptyChange() error {
	fmt.Println("No changes so making an empty commit to force rebuild")
	_, err := h.commit.Commit("Add empty commit to force rebuild on continuous pipe")
	return err
}
