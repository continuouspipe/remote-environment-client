package cmd

import (
	"github.com/spf13/cobra"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/git"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Create/Update the remote environment",
	Long: `The build command will push changes the branch you have checked out locally to your remote 
environment branch. ContinuousPipe will then build the environment. You can use the 
https://ui.continuouspipe.io/ to see when the environment has finished building and to 
find its IP address.`,
	Run: func(cmd *cobra.Command, args []string) {
		settings := config.NewApplicationSettings()
		validator := config.NewMandatoryChecker()
		validateConfig(validator, settings)

		handler := &BuildHandle{cmd}
		gitCommitTrigger := git.NewGitCommitTrigger()
		handler.Handle(args, settings, gitCommitTrigger)
	},
}

func init() {
	RootCmd.AddCommand(buildCmd)
}

type BuildHandle struct {
	Command *cobra.Command
}

func (h *BuildHandle) Handle(args []string, settings config.Reader, commitTrigger git.CommitTrigger) {
	remoteName := settings.GetString(config.RemoteName)
	remoteBranch := settings.GetString(config.RemoteBranch)
	commitTrigger.PushEmptyCommit(remoteName, remoteBranch)
}
