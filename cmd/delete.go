package cmd

import (
	"fmt"
	"io"
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
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/session"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	kubectlcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

var (
	deleteLong    = templates.LongDesc(msgs.DeleteCommandLongDescription)
	deleteExample = templates.Examples(msgs.DeleteCommandExampleDescription)
)

//DeleteCmdName is the command name identifier
const DeleteCmdName = "delete"

//NewDeleteCmd returns a new command that wraps the kubectl delete command
//it finds the target pod and pass the arguments to the wrapped command
func NewDeleteCmd() *cobra.Command {
	settings := config.C

	handler := &DeletePodCmdHandle{}
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	handler.writer = os.Stdout

	command := &cobra.Command{
		Use:     fmt.Sprintf("%s ([-f FILENAME] | TYPE [(NAME | -l label | --all)])", DeleteCmdName),
		Short:   msgs.DeleteCommandShortDescription,
		Long:    deleteLong,
		Example: deleteExample,
		Run: func(cmd *cobra.Command, args []string) {
			remoteCommand := remotecplogs.NewRemoteCommand(DeleteCmdName, os.Args)
			cs := session.NewCommandSession().Start()

			//validate the configuration file
			missingSettings, ok := config.C.Validate()
			if ok == false {
				reason := fmt.Sprintf(msgs.InvalidConfigSettings, missingSettings)
				err := remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.Ended(http.StatusBadRequest, reason, "", *cs))
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cs, err)
				cperrors.ExitWithMessage(reason)
			}

			handler.Complete(args, settings)
			err := handler.Validate()
			if err != nil {
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cs, err)
				cperrors.ExitWithMessage(err.Error())
			}

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

	//kubectl delete cmd options
	command.Flags().StringVarP(&handler.options.selector, "selector", "l", "", "Selector (label query) to filter on.")
	command.Flags().BoolVar(&handler.options.all, "all", false, "[-all] to select all the specified resources.")
	command.Flags().BoolVar(&handler.options.ignoreNotFound, "ignore-not-found", false, "Treat \"resource not found\" as a successful delete. Defaults to \"true\" when --all is specified.")
	command.Flags().BoolVar(&handler.options.cascade, "cascade", true, "If true, cascade the deletion of the resources managed by this resource (e.g. Pods created by a ReplicationController).  Default true.")
	command.Flags().IntVar(&handler.options.gracePeriod, "grace-period", -1, "Period of time in seconds given to the resource to terminate gracefully. Ignored if negative.")
	command.Flags().BoolVar(&handler.options.now, "now", false, "If true, resources are signaled for immediate shutdown (same as --grace-period=1).")
	command.Flags().BoolVar(&handler.options.force, "force", false, "Immediate deletion of some resources may result in inconsistency or data loss and requires confirmation.")
	command.Flags().DurationVar(&handler.options.timeout, "timeout", 0, "The length of time to wait before giving up on a delete, zero means determine a timeout from the size of the object")

	//cp tool added options
	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	//cp-remote specific cmd options
	command.Flags().StringVarP(&handler.options.environment, config.KubeEnvironmentName, "e", environment, "The full remote environment name")

	return command
}

//DeletePodCmdHandle holds the information required to delete a pod
type DeletePodCmdHandle struct {
	kubeCtlInit kubectlapi.KubeCtlInitializer
	writer      io.Writer
	options     deletePodCmdOptions
	argsIn      []string
}

type deletePodCmdOptions struct {
	environment, selector                    string
	all, ignoreNotFound, cascade, now, force bool
	gracePeriod                              int
	timeout                                  time.Duration
}

// Complete verifies command line arguments and loads data from the command environment
func (h *DeletePodCmdHandle) Complete(argsIn []string, settings *config.Config) {
	if h.options.environment == "" {
		h.options.environment = settings.GetStringQ(config.KubeEnvironmentName)
	}
	h.argsIn = argsIn
}

// Validate checks that the provided exec options are specified.
func (h *DeletePodCmdHandle) Validate() error {
	if len(strings.Trim(h.options.environment, " ")) == 0 {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.EnvironmentSpecifiedEmpty).String())
	}
	return nil
}

// Handle set the flags in the kubeclt logs handle command and executes it
func (h *DeletePodCmdHandle) Handle(args []string) (suggestion string, err error) {
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionGetSettingsError, session.CurrentSession.SessionID), err
	}

	clientConfig := kubectlapi.GetNonInteractiveDeferredLoadingClientConfig(user, apiKey, addr, h.options.environment)
	kubeCmdDelete := kubectlcmd.NewCmdDelete(kubectlcmdutil.NewFactory(clientConfig), os.Stdout)

	kubeCmdDelete.Flags().Set("all", strconv.FormatBool(h.options.all))
	kubeCmdDelete.Flags().Set("cascade", strconv.FormatBool(h.options.cascade))
	kubeCmdDelete.Flags().Set("force", strconv.FormatBool(h.options.force))
	kubeCmdDelete.Flags().Set("grace-period", strconv.Itoa(h.options.gracePeriod))
	kubeCmdDelete.Flags().Set("ignore-not-found", strconv.FormatBool(h.options.ignoreNotFound))
	kubeCmdDelete.Flags().Set("now", strconv.FormatBool(h.options.now))
	kubeCmdDelete.Flags().Set("selector", h.options.selector)
	kubeCmdDelete.Flags().Set("timeout", h.options.timeout.String())

	err = kubectlcmdutil.ValidateOutputArgs(kubeCmdDelete)
	if err != nil {
		return err.Error(), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "kubernetes did not validate the arguments provided").String())
	}
	err = kubectlcmd.RunDelete(kubectlcmdutil.NewFactory(clientConfig), os.Stdout, kubeCmdDelete, args, &resource.FilenameOptions{})
	if err != nil {
		return err.Error(), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "kubernetes failed to delete resource").String())
	}

	return "", nil
}
