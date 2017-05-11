package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/initialization"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/session"
	"github.com/pkg/errors"
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
		Short:   "Create/Update the remote environment",
		Long: `The build command will push changes the branch you have checked out locally to your remote
environment branch. ContinuousPipe will then build the environment. You can use the
https://ui.continuouspipe.io/ to see when the environment has finished building and to
find its IP address.`,
		Run: func(cmd *cobra.Command, args []string) {
			remoteCommand := cplogs.NewRemoteCommand(BuildCmdName, args)
			cmdSession := session.NewCommandSession().Start()

			//validate the configuration file
			missingSettings, ok := config.C.Validate()
			if ok == false {
				reason := fmt.Sprintf(msgs.InvalidConfigSettings, missingSettings)
				cplogs.NewRemoteCommandSender().Send(*remoteCommand.Ended(http.StatusBadRequest, reason, "", *cmdSession))
				cperrors.ExitWithMessage(reason)
			}

			//call the build handler
			err := handler.Handle()
			if err != nil {
				code, reason, stack := cperrors.FindCause(err)
				cplogs.NewRemoteCommandSender().Send(*remoteCommand.Ended(code, reason, stack, *cmdSession))
				cperrors.ExitWithMessage(err.Error())
			}
			cplogs.NewRemoteCommandSender().Send(*remoteCommand.EndedOk(*cmdSession))
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
func (h *BuildHandle) Handle() error {
	err := h.triggerBuild.Handle()
	if err != nil {
		return errors.Wrapf(err, msgs.SuggestionTriggerBuildFailed, session.CurrentSession.SessionID)
	}
	err = h.waitForEnvironmentReady.Handle()
	if err != nil {
		return errors.Wrapf(err, msgs.SuggestionWaitForEnvironmentReadyFailed, session.CurrentSession.SessionID)
	}

	h.config.Set(config.InitStatus, initStateCompleted)
	err = h.config.Save(config.AllConfigTypes)
	if err != nil {
		return errors.Wrapf(err, msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID)
	}

	apiKey := h.config.GetStringQ(config.ApiKey)
	h.api.SetAPIKey(apiKey)

	remoteEnvID := h.config.GetStringQ(config.RemoteEnvironmentId)
	flowID := h.config.GetStringQ(config.FlowId)

	remoteEnv, err := h.api.GetRemoteEnvironmentStatus(flowID, remoteEnvID)
	if err != nil {
		return errors.Wrapf(err, msgs.SuggestionGetEnvironmentStatusFailed, session.CurrentSession.SessionID)
	}

	fmt.Fprintf(h.stdout, "\n\n# Get started !\n")
	fmt.Fprintln(h.stdout, "You can now run `cp-remote watch` to watch your local changes with the deployed environment ! Your deployed environment can be found at this address:")
	fmt.Fprintf(h.stdout, "\n\nCheckout the documentation at https://docs.continuouspipe.io/remote-development/ \n")
	cpapi.PrintPublicEndpoints(h.stdout, remoteEnv.PublicEndpoints)

	return nil
}
