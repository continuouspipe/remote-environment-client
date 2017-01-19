package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/git"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy the remote environment",
	Long: `The destroy command will delete the remote branch used for your remote
environment, ContinuousPipe will then remove the environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		settings := config.NewApplicationSettings()
		validator := config.NewMandatoryChecker()
		validateConfig(validator, settings)

		fmt.Println("Destroying remote environment")
		fmt.Println("Deleting remote branch")

		gitBranchRemover := git.NewGitBranchRemover()
		handler := &DestroyHandle{cmd}
		handler.Handle(args, settings, gitBranchRemover)

		fmt.Println("Continuous Pipe will now destroy the remote environment")
	},
}

func init() {
	RootCmd.AddCommand(destroyCmd)
}

type DestroyHandle struct {
	Command *cobra.Command
}

func (h *DestroyHandle) Handle(args []string, settings config.Reader, branchRemover git.BranchRemover) {
	remoteName := settings.GetString(config.RemoteName)
	remoteBranch := settings.GetString(config.RemoteBranch)
	branchRemover.Delete(remoteName, remoteBranch)
}
