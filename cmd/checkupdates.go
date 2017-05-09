package cmd

import (
	"github.com/continuouspipe/remote-environment-client/update"
	"github.com/spf13/cobra"
)

func NewCheckUpdatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "checkupdates",
		Aliases: []string{"ckup"},
		Short:   "Check for latest version",
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			handler := &CheckUpdates{cmd}
			handler.Handle(args)
		},
	}
}

type CheckUpdates struct {
	Command *cobra.Command
}

func (h *CheckUpdates) Handle(args []string) {
	err := update.CheckForLatestVersion()
	if err != nil {
		//TODO: Send error log to Sentry
		//TODO: Log err
		//TODO: Print user friendly error that explains what happened and what to do next
	}
}
