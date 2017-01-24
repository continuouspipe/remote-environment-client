package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/git"
	"github.com/spf13/cobra"
	"strings"
)

func NewDestroyCmd() *cobra.Command {
	settings := config.NewApplicationSettings()
	handler := &DestroyHandle{}
	command := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy the remote environment",
		Long: `The destroy command will delete the remote branch used for your remote
environment, ContinuousPipe will then remove the environment.`,
		Run: func(cmd *cobra.Command, args []string) {
			validator := config.NewMandatoryChecker()
			validateConfig(validator, settings)

			fmt.Println("Destroying remote environment")
			fmt.Println("Deleting remote branch")

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle())

			fmt.Println("Continuous Pipe will now destroy the remote environment")
		},
	}
	return command
}

type DestroyHandle struct {
	Command      *cobra.Command
	remoteName   string
	remoteBranch string
}

// Complete verifies command line arguments and loads data from the command environment
func (h *DestroyHandle) Complete(cmd *cobra.Command, argsIn []string, settingsReader config.Reader) error {
	h.Command = cmd
	h.remoteName = settingsReader.GetString(config.RemoteName)
	h.remoteBranch = settingsReader.GetString(config.RemoteBranch)
	return nil
}

// Validate checks that the provided destroy options are specified.
func (h *DestroyHandle) Validate() error {
	if len(strings.Trim(h.remoteName, " ")) == 0 {
		return fmt.Errorf("the remote name specified is invalid")
	}
	if len(strings.Trim(h.remoteBranch, " ")) == 0 {
		return fmt.Errorf("the remote branch specified is invalid")
	}
	return nil
}

func (h *DestroyHandle) Handle() error {
	push := git.NewPush()
	_, err := push.DeleteRemote(h.remoteName, h.remoteBranch)
	return err
}
