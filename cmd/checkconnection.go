package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	remotecplogs "github.com/continuouspipe/remote-environment-client/cplogs/remote"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/session"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl"
)

//NewListPodsCmd return a cobra cmd alias of NewCheckConnectionCmd
func NewListPodsCmd() *cobra.Command {
	ck := NewCheckConnectionCmd()
	ck.Use = "pods"
	ck.Short = "Lists the pods in the remote environment (alias for checkconnection)"
	ck.Aliases = []string{"po"}
	return ck
}

//CheckConnectionCmdName is the command name identifier
const CheckConnectionCmdName = "checkconnection"

//NewCheckConnectionCmd return a new cobra command that on run executes the CheckConnectionHandle
func NewCheckConnectionCmd() *cobra.Command {
	settings := config.C
	handler := &CheckConnectionHandle{}
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	command := &cobra.Command{
		Use:     CheckConnectionCmdName,
		Aliases: []string{"ck"},
		Short:   msgs.CheckConnectionCommandShortDescription,
		Long:    msgs.CheckConnectionCommandLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			remoteCommand := remotecplogs.NewRemoteCommand(CheckConnectionCmdName, os.Args)
			cs := session.NewCommandSession().Start()

			//validate the configuration file
			missingSettings, ok := config.C.Validate()
			if ok == false {
				reason := fmt.Sprintf(msgs.InvalidConfigSettings, missingSettings)
				err := remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.Ended(http.StatusBadRequest, reason, "", *cs))
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cs, err)
				cperrors.ExitWithMessage(reason)
			}

			podsFinder := pods.NewKubePodsFind()
			handler.Complete(cmd, args, settings)
			err := handler.Validate()
			if err != nil {
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cs, err)
				cperrors.ExitWithMessage(err.Error())
			}

			//call the command handler
			suggestion, err := handler.Handle(args, podsFinder)
			if err != nil {
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cs, err)
				cperrors.ExitWithMessage(suggestion)
			}

			//send the command metrics
			err = remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.EndedOk(*cs))
			if err != nil {
				cplogs.V(4).Infof(remotecplogs.ErrorFailedToSendDataToLoggingAPI)
				cplogs.Flush()
			}
		},
	}

	command.PersistentFlags().StringVarP(&handler.Environment, config.KubeEnvironmentName, "e", settings.GetStringQ(config.KubeEnvironmentName), "The full remote environment name")
	return command
}

//CheckConnectionHandle holds the command handler dependencies
type CheckConnectionHandle struct {
	Command     *cobra.Command
	Environment string
	kubeCtlInit kubectlapi.KubeCtlInitializer
}

//Complete verifies command line arguments and loads data from the command environment
func (h *CheckConnectionHandle) Complete(cmd *cobra.Command, argsIn []string, setting *config.Config) {
	h.Command = cmd
	if h.Environment == "" {
		h.Environment = setting.GetStringQ(config.KubeEnvironmentName)
	}
}

//Validate checks that the provided checkconnection options are specified.
func (h *CheckConnectionHandle) Validate() error {
	if len(strings.Trim(h.Environment, " ")) == 0 {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.EnvironmentSpecifiedEmpty).String())
	}
	return nil
}

//Handle finds the pods and prints them
func (h *CheckConnectionHandle) Handle(args []string, podsFinder pods.Finder) (suggestion string, err error) {
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionGetSettingsError, session.CurrentSession.SessionID), err
	}

	fmt.Println(fmt.Sprintf(msgs.CheckingConnectionForEnvironment, h.Environment))

	podsList, err := podsFinder.FindAll(user, apiKey, addr, h.Environment)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionFindPodsFailed, session.CurrentSession.SessionID), err
	}

	if len(podsList.Items) > 0 {
		printer := kubectl.NewHumanReadablePrinter(kubectl.PrintOptions{
			ColumnLabels:  []string{},
			Wide:          true,
			WithNamespace: true,
		})
		printer.EnsurePrintWithKind(podsList.Kind)
		color.Green(msgs.PodsFoundCount, len(podsList.Items))
		printer.PrintObj(podsList, os.Stdout)
	} else {
		color.Red(msgs.PodsNotFound)
	}

	return "", nil
}
