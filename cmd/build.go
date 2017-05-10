package cmd

import (
	"fmt"
	"io"
	"os"

	"net/http"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/errors"
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
	handler.api = cpapi.NewCpApi()
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
			valid, missing := config.C.Validate()
			if valid == false {
				reason := fmt.Sprintf(msgs.InvalidConfigSettings, missing)
				cplogs.NewRemoteCommandSender().Send(*remoteCommand.Ended(http.StatusBadRequest, reason, *cmdSession))
				errors.ExitWithMessage(reason)
			}

			//call the build handler
			err := handler.Handle()

			checkErr(err)
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
	api                     cpapi.CpApiProvider
}

//Handle performs the 2 init stages that trigger that build and wait for the environment to be ready
func (h *BuildHandle) Handle() error {
	err := h.triggerBuild.Handle()
	if err != nil {
		//TODO: Send error log to Sentry
		//TODO: Log err
		//TODO: Print user friendly error that explains what happened and what to do next
		return err
	}
	err = h.waitForEnvironmentReady.Handle()
	if err != nil {
		//TODO: Send error log to Sentry
		//TODO: Log err
		//TODO: Print user friendly error that explains what happened and what to do next
		return err
	}

	h.config.Set(config.InitStatus, initStateCompleted)
	err = h.config.Save(config.AllConfigTypes)
	if err != nil {
		//TODO: Send error log to Sentry
		//TODO: Log err
		//TODO: Print user friendly error that explains what happened and what to do next
		return err
	}

	apiKey, err := h.config.GetString(config.ApiKey)
	if err != nil {
		return err
	}
	remoteEnvID, err := h.config.GetString(config.RemoteEnvironmentId)
	if err != nil {
		return err
	}
	flowID, err := h.config.GetString(config.FlowId)
	if err != nil {
		return err
	}

	h.api.SetApiKey(apiKey)

	remoteEnv, err := h.api.GetRemoteEnvironmentStatus(flowID, remoteEnvID)
	if err != nil {
		//TODO: Send error log to Sentry
		//TODO: Log err
		//TODO: Print user friendly error that explains what happened and what to do next
		return err
	}

	fmt.Fprintf(h.stdout, "\n\n# Get started !\n")
	fmt.Fprintln(h.stdout, "You can now run `cp-remote watch` to watch your local changes with the deployed environment ! Your deployed environment can be found at this address:")
	fmt.Fprintf(h.stdout, "\n\nCheckout the documentation at https://docs.continuouspipe.io/remote-development/ \n")
	cpapi.PrintPublicEndpoints(h.stdout, remoteEnv.PublicEndpoints)

	return nil
}
