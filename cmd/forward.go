package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/client/unversioned/portforward"
	"k8s.io/kubernetes/pkg/client/unversioned/remotecommand"
	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	kubectlcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

var (
	portforwardExample = fmt.Sprintf(msgs.PortFowardCommandExampleDescription, config.AppName)
)

const ForwardCmdName = "forward"

func NewForwardCmd() *cobra.Command {
	settings := config.C
	handler := &ForwardHandle{}
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	command := &cobra.Command{
		Use:     ForwardCmdName,
		Aliases: []string{"fo"},
		Short:   msgs.PortForwardCommandShortDescription,
		Long:    msgs.PortForwardCommandLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			remoteCommand := remotecplogs.NewRemoteCommand(ForwardCmdName, os.Args)
			cmdSession := session.NewCommandSession().Start()

			//validate the configuration file
			missingSettings, ok := config.C.Validate()
			if ok == false {
				reason := fmt.Sprintf(msgs.InvalidConfigSettings, missingSettings)
				err := remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.Ended(http.StatusBadRequest, reason, "", *cmdSession))
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cmdSession, err)
				cperrors.ExitWithMessage(reason)
			}

			//complete the option and validate them
			handler.Complete(cmd, args, settings)
			err := handler.Validate()
			if err != nil {
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cmdSession, err)
				cperrors.ExitWithMessage(err.Error())
			}

			//call the command handler
			suggestion, err := handler.Handle()
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
		Example: portforwardExample,
	}

	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	service, err := settings.GetString(config.Service)
	checkErr(err)

	command.PersistentFlags().StringVarP(&handler.Environment, config.KubeEnvironmentName, "e", environment, "The full remote environment name")
	command.PersistentFlags().StringVarP(&handler.Service, config.Service, "s", service, "The service to use (e.g.: web, mysql)")
	return command
}

type ForwardHandle struct {
	Command     *cobra.Command
	ports       []string
	podsFinder  pods.Finder
	podsFilter  pods.Filter
	Environment string
	Service     string
	kubeCtlInit kubectlapi.KubeCtlInitializer
}

// Complete verifies command line arguments and loads data from the command environment
func (h *ForwardHandle) Complete(cmd *cobra.Command, argsIn []string, settings *config.Config) {
	h.Command = cmd
	h.ports = argsIn

	h.podsFinder = pods.NewKubePodsFind()
	h.podsFilter = pods.NewKubePodsFilter()

	if h.Environment == "" {
		h.Environment = settings.GetStringQ(config.KubeEnvironmentName)
	}
	if h.Service == "" {
		h.Service = settings.GetStringQ(config.Service)
	}
}

// Validate checks that the provided forward options are specified.
func (h *ForwardHandle) Validate() error {
	if len(h.ports) == 0 {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "at least 1 PORT is required for port-forward").String())
	}
	if len(strings.Trim(h.Environment, " ")) == 0 {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.EnvironmentSpecifiedEmpty).String())
	}
	if len(strings.Trim(h.Service, " ")) == 0 {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.ServiceSpecifiedEmpty).String())
	}
	return nil
}

func (h *ForwardHandle) Handle() (suggestion string, err error) {
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionGetSettingsError, session.CurrentSession.SessionID), err
	}

	allPods, err := h.podsFinder.FindAll(user, apiKey, addr, h.Environment)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionFindPodsFailed, session.CurrentSession.SessionID), err
	}

	pod := h.podsFilter.List(*allPods).ByService(h.Service).ByStatus("Running").ByStatusReason("Running").First()
	if pod == nil {
		return fmt.Sprintf(msgs.SuggestionRunningPodNotFound, h.Service, h.Environment, config.AppName, "bash", session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, fmt.Sprintf(msgs.NoActivePodsFoundForSpecifiedServiceName, h.Service)).String())
	}

	cplogs.V(5).Infof("setting up forwarding for target pod %s and ports %s", pod.GetName(), h.ports)
	cplogs.Flush()

	clientConfig := kubectlapi.GetNonInteractiveDeferredLoadingClientConfig(user, apiKey, addr, h.Environment)
	kubeCmdPortForward := kubectlcmd.NewCmdPortForward(kubectlcmdutil.NewFactory(clientConfig), os.Stdout, os.Stderr)

	opts := &kubectlcmd.PortForwardOptions{
		PortForwarder: &defaultPortForwarder{
			cmdOut: os.Stdout,
			cmdErr: os.Stderr,
		},
	}

	if err := opts.Complete(kubectlcmdutil.NewFactory(clientConfig), kubeCmdPortForward, append([]string{pod.GetName()}, h.ports...), os.Stdout, os.Stderr); err != nil {
		return err.Error(), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "kubernetes lib returned an error completing port forward options").String())
	}
	if err := opts.Validate(); err != nil {
		return err.Error(), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "kubernetes lib returned an error validating port forward options").String())
	}
	if err := opts.RunPortForward(); err != nil {
		return err.Error(), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "kubernetes lib returned an error executing port forward options").String())
	}

	return "", nil
}

type defaultPortForwarder struct {
	cmdOut, cmdErr io.Writer
}

func (f *defaultPortForwarder) ForwardPorts(method string, url *url.URL, opts kubectlcmd.PortForwardOptions) error {
	dialer, err := remotecommand.NewExecutor(opts.Config, method, url)
	if err != nil {
		return err
	}
	fw, err := portforward.New(dialer, opts.Ports, opts.StopChannel, opts.ReadyChannel, f.cmdOut, f.cmdErr)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}
