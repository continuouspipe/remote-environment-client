package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	remotecplogs "github.com/continuouspipe/remote-environment-client/cplogs/remote"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/session"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	kubectlcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

//LogsCmdName is the command name identifier
const LogsCmdName = "logs"

func NewLogsCmd() *cobra.Command {
	settings := config.C
	handler := &LogsCmdHandle{}
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	command := &cobra.Command{
		Use:     LogsCmdName,
		Aliases: []string{"lo"},
		Short:   msgs.LogsCommandShortDescription,
		Long:    msgs.LogsCommandLongDescription,
		Example: fmt.Sprintf(msgs.LogsCommandExampleDescription, config.AppName),
		Run: func(cmd *cobra.Command, args []string) {
			remoteCommand := remotecplogs.NewRemoteCommand(LogsCmdName, os.Args)
			cmdSession := session.NewCommandSession().Start()

			//validate the configuration file
			missingSettings, ok := config.C.Validate()
			if ok == false {
				reason := fmt.Sprintf(msgs.InvalidConfigSettings, missingSettings)
				err := remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.Ended(http.StatusBadRequest, reason, "", *cmdSession))
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cmdSession, err)
				cperrors.ExitWithMessage(reason)
			}

			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()

			handler.Complete(cmd, args, settings)
			err := handler.Validate()
			if err != nil {
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cmdSession, err)
				cperrors.ExitWithMessage(err.Error())
			}
			//call the command handler
			suggestion, err := handler.Handle(args, podsFinder, podsFilter)
			if err != nil {
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cmdSession, err)
				cperrors.ExitWithMessage(suggestion)
			}

			//send the command metrics
			err = remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.EndedOk(*cmdSession))
			if err != nil {
				cplogs.V(4).Infof(remotecplogs.ErrorFailedToSendDataToLoggingAPI)
				cplogs.Flush()
			}

		},
	}
	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	service, err := settings.GetString(config.Service)
	checkErr(err)

	//used to find the targed pod
	command.PersistentFlags().StringVarP(&handler.environment, config.KubeEnvironmentName, "e", environment, "The full remote environment name")
	command.PersistentFlags().StringVarP(&handler.service, config.Service, "s", service, "The service to use (e.g.: web, mysql)")

	command.PersistentFlags().DurationVar(&handler.since, "since", 0, "Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.")
	command.PersistentFlags().Int64Var(&handler.tail, "tail", -1, "Lines of recent log file to display. Defaults to -1, showing all log lines.")
	command.PersistentFlags().BoolVarP(&handler.follow, "follow", "f", false, "Specify if the logs should be streamed.")
	command.PersistentFlags().BoolVarP(&handler.previous, "previous", "p", false, "If true, print the logs for the previous instance of the container in a pod if it exists.")
	return command
}

type LogsCmdHandle struct {
	environment string
	service     string
	username    string
	apiKey      string
	conf        config.ConfigProvider
	kubeCtlInit kubectlapi.KubeCtlInitializer
	since       time.Duration
	tail        int64
	follow      bool
	previous    bool
}

// Complete verifies command line arguments and loads data from the command environment
func (h *LogsCmdHandle) Complete(cmd *cobra.Command, argsIn []string, settings *config.Config) {
	if h.environment == "" {
		h.environment = settings.GetStringQ(config.KubeEnvironmentName)
	}
	if h.service == "" {
		h.service = settings.GetStringQ(config.Service)
	}
	if h.username == "" {
		h.username = settings.GetStringQ(config.Username)
	}
	if h.apiKey == "" {
		h.apiKey = settings.GetStringQ(config.ApiKey)
	}
}

// Validate checks that the provided exec options are specified.
func (h *LogsCmdHandle) Validate() error {
	if len(strings.Trim(h.environment, " ")) == 0 {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.EnvironmentSpecifiedEmpty).String())
	}
	if len(strings.Trim(h.service, " ")) == 0 {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.ServiceSpecifiedEmpty).String())
	}
	return nil
}

// Handle set the flags in the kubeclt logs handle command and executes it
func (h *LogsCmdHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter) (suggestion string, err error) {
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionGetSettingsError, session.CurrentSession.SessionID), err
	}

	allPods, err := podsFinder.FindAll(user, apiKey, addr, h.environment)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionFindPodsFailed, session.CurrentSession.SessionID), err
	}

	pod := podsFilter.List(*allPods).ByService(h.service).ByStatus("Running").ByStatusReason("Running").First()
	if pod == nil {
		return fmt.Sprintf(msgs.SuggestionRunningPodNotFound, h.service, h.environment, config.AppName, "bash", session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, fmt.Sprintf(msgs.NoActivePodsFoundForSpecifiedServiceName, h.service)).String())
	}

	cplogs.V(5).Infof("getting container logs for environment %s, pod %s", h.environment, pod.GetName())
	cplogs.Flush()

	clientConfig := kubectlapi.GetNonInteractiveDeferredLoadingClientConfig(user, apiKey, addr, h.environment)
	f := kubectlcmdutil.NewFactory(clientConfig)
	kubeCmdLogs := kubectlcmd.NewCmdLogs(f, os.Stdout)

	kubeCmdLogs.Flags().Set("since", h.since.String())
	kubeCmdLogs.Flags().Set("tail", strconv.FormatInt(h.tail, 10))
	kubeCmdLogs.Flags().Set("follow", strconv.FormatBool(h.follow))
	kubeCmdLogs.Flags().Set("previous", strconv.FormatBool(h.previous))

	args = append([]string{pod.GetName()}, args...)

	o := &kubectlcmd.LogsOptions{}
	o.Complete(f, os.Stdout, kubeCmdLogs, args)
	if err := o.Validate(); err != nil {
		return "", kubectlcmdutil.UsageError(kubeCmdLogs, err.Error())
	}
	_, err = o.RunLogs()
	if err != nil {
		return "", err
	}

	return "", nil
}
