package cmd

import (
	"fmt"

	"github.com/continuouspipe/remote-environment-client/cplogs"
	remotecplogs "github.com/continuouspipe/remote-environment-client/cplogs/remote"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/session"
	"github.com/continuouspipe/remote-environment-client/update"
	"github.com/spf13/cobra"
)

const CheckUpdatesCmdName = "checkupdates"

func NewCheckUpdatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     CheckUpdatesCmdName,
		Aliases: []string{"ckup"},
		Short:   "Check for latest version",
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			remoteCommand := remotecplogs.NewRemoteCommand(CheckUpdatesCmdName, args)
			cs := session.NewCommandSession().Start()

			handler := &CheckUpdates{cmd}
			suggestion, err := handler.Handle(args)
			if err != nil {
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cs, err)
				cperrors.ExitWithMessage(suggestion)
			}
			err = remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.EndedOk(*cs))
			if err != nil {
				cplogs.V(4).Infof(remotecplogs.ErrorFailedToSendDataToLoggingAPI)
				cplogs.Flush()
			}
		},
	}
}

type CheckUpdates struct {
	Command *cobra.Command
}

func (h *CheckUpdates) Handle(args []string) (suggestion string, err error) {
	err = update.CheckForLatestVersion()
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionCheckForLatestVersionFailed, session.CurrentSession.SessionID), err
	}
	return "", nil
}
