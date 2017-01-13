package cmd

import (
	"github.com/spf13/cobra"
	"github.com/continuouspipe/remote-environment-client/update"
)

var checkUpdates = &cobra.Command{
	Use:   "checkupdates",
	Short: "Check for latest version",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		handler := &CheckUpdates{cmd}
		handler.Handle(args)
	},
}

func init() {
	RootCmd.AddCommand(checkUpdates)
}

type CheckUpdates struct {
	Command *cobra.Command
}

func (h *CheckUpdates) Handle(args []string) {
	update.CheckForLatestVersion()
}
