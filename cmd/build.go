package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	remotecplogs "github.com/continuouspipe/remote-environment-client/cplogs/remote"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/initialization"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/session"
	"github.com/spf13/cobra"
)

//BuildCmdName is the name identifier for the build command
const BuildCmdName = "build"

//NewBuildCmd return a new cobra command that handles the build
func NewBuildCmd() *cobra.Command {
	handler := &BuildHandle{}
	handler.stdout = os.Stdout
	handler.config = config.C
	handler.triggerBuild = newTriggerBuild()
	handler.waitForEnvironmentReady = newWaitEnvironmentReady()
	handler.api = cpapi.NewCpAPI()
	command := &cobra.Command{
		Use:     BuildCmdName,
		Aliases: []string{"bu"},
		Short:   msgs.BuildCommandShortDescription,
		Long:    msgs.BuildCommandLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			remoteCommand := remotecplogs.NewRemoteCommand(BuildCmdName, os.Args)
			cs := session.NewCommandSession().Start()

			//validate the configuration file
			missingSettings, ok := config.C.Validate()
			if ok == false {
				reason := fmt.Sprintf(msgs.InvalidConfigSettings, missingSettings)
				err := remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.Ended(http.StatusBadRequest, reason, "", *cs))
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cs, err)
				cperrors.ExitWithMessage(reason)
			}

			//call the build handler
			suggestion, err := handler.Handle()
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
	return command
}

//BuildHandle holds the dependencies of the build handler
type BuildHandle struct {
	config                  config.ConfigProvider
	triggerBuild            initialization.InitState
	waitForEnvironmentReady initialization.InitState
	stdout                  io.Writer
	api                     cpapi.DataProvider
}

//Handle performs the 2 init stages that trigger that build and wait for the environment to be ready
func (h *BuildHandle) Handle() (suggestion string, err error) {
	suggestion, err = h.triggerBuild.Handle()
	if err != nil {
		return suggestion, err
	}
	suggestion, err = h.waitForEnvironmentReady.Handle()
	if err != nil {
		return suggestion, err
	}

	h.config.Set(config.InitStatus, initStateCompleted)
	err = h.config.Save(config.AllConfigTypes)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID), err
	}

	apiKey := h.config.GetStringQ(config.ApiKey)
	h.api.SetAPIKey(apiKey)

	remoteEnvID := h.config.GetStringQ(config.RemoteEnvironmentId)
	flowID := h.config.GetStringQ(config.FlowId)

	remoteEnv, err := h.api.GetRemoteEnvironmentStatus(flowID, remoteEnvID)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionGetEnvironmentStatusFailed, session.CurrentSession.SessionID), err
	}

	fmt.Fprintf(h.stdout, fmt.Sprintf("%s\n", msgs.GetStarted))
	cpapi.PrintPublicEndpoints(h.stdout, remoteEnv.PublicEndpoints)
	fmt.Fprintf(h.stdout, "\n\n%s\n", msgs.CheckDocumentation)

	return "", nil
}
