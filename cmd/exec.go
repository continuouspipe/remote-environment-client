package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
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
	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	kubectlcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

//ExecCmdName is the name identifier for the exec command
const ExecCmdName = "exec"

//NewExecCmd return a cobra command struct pointer which on Run, if required it prepares the config so we can reach the pod and
//then uses a command handler to execute the command specified in the arguments
func NewExecCmd() *cobra.Command {
	handler := newExecHandle()

	var interactive bool
	var flowID string

	bashcmd := &cobra.Command{
		Use:     ExecCmdName,
		Aliases: []string{"ex"},
		Short:   msgs.ExecCommandShortDescription,
		Long:    msgs.ExecCommandLongDescription,
		Example: fmt.Sprintf(msgs.ExecCommandExampleDescription, config.AppName),
		Run: func(cmd *cobra.Command, args []string) {
			remoteCommand := remotecplogs.NewRemoteCommand(ExecCmdName, args)
			cmdSession := session.NewCommandSession().Start()

			suggestion, err := RunExec(handler, interactive, flowID, args)
			if err != nil {
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cmdSession, err)
				cperrors.ExitWithMessage(suggestion)
			}

			err = remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.EndedOk(*cmdSession))
			if err != nil {
				cplogs.V(4).Infof(remotecplogs.ErrorFailedToSendDataToLoggingAPI)
				cplogs.Flush()
			}
		},
	}

	bashcmd.PersistentFlags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode allows to target a different environment")

	bashcmd.PersistentFlags().StringVarP(&handler.environment, config.KubeEnvironmentName, "e", "", "The full remote environment name")
	bashcmd.PersistentFlags().StringVarP(&handler.service, config.Service, "s", "", "The service to use (e.g.: web, mysql)")
	bashcmd.PersistentFlags().StringVarP(&flowID, config.FlowId, "f", "", "The flow to use")

	return bashcmd
}

func RunExec(handler *execHandle, interactive bool, flowID string, args []string) (suggestion string, err error) {
	settings := config.C

	podsFinder := pods.NewKubePodsFind()
	podsFilter := pods.NewKubePodsFilter()

	defaultEnvironment := settings.GetStringQ(config.KubeEnvironmentName)
	defaultService := settings.GetStringQ(config.Service)

	if interactive {
		cplogs.V(5).Infoln("exec in interactive mode")
		//make sure config has an api key and a cp user set
		initInteractiveH := NewInitInteractiveHandler(false)
		initInteractiveH.SetWriter(ioutil.Discard)
		err := initInteractiveH.Complete(args)
		if err != nil {
			return "", err
		}
		err = initInteractiveH.Validate()
		if err != nil {
			return "", err
		}
		err = initInteractiveH.Handle()
		if err != nil {
			return err.Error(), err
		}

		if flowID == "" && handler.environment == "" && handler.service == "" {
			//guide the user to choose the right pod they want to target
			questioner := cpapi.NewMultipleChoiceCpEntityQuestioner()
			questioner.SetAPIKey(settings.GetStringQ(config.ApiKey))
			_, flow, environment, pod, suggestion, err := questioner.WhichEntities()
			if err != nil {
				return suggestion, err
			}

			handler.environment = environment.Identifier
			handler.service = pod.Name
			flowID = flow.UUID

			suggestedFlags := color.GreenString("-i -e %s -f %s -s %s", environment.Identifier, flow.UUID, pod.Name)
			fmt.Printf(fmt.Sprintf("\n\n%s\n", msgs.InteractiveModeSuggestingFlags), suggestedFlags)
		}

		//alter the configuration so that we connect to the flow and environment specified by the user
		suggestion, err = newInteractiveModeH().findTargetClusterAndApplyToConfig(flowID, handler.environment)
		if err != nil {
			return suggestion, err
		}
	} else {
		//apply default values
		if handler.environment == "" {
			handler.environment = defaultEnvironment
		}
		if handler.service == "" {
			handler.service = defaultService
		}
	}

	handler.complete(args, settings)
	err = handler.validate()
	if err != nil {
		return err.Error(), err
	}
	suggestion, err = handler.handle(podsFinder, podsFilter)
	if err != nil {
		return suggestion, err
	}
	return "", nil
}

type execHandle struct {
	args        []string
	config      config.ConfigProvider
	environment string
	service     string
	kubeCtlInit kubectlapi.KubeCtlInitializer
}

func newExecHandle() *execHandle {
	p := &execHandle{}
	p.config = config.C
	p.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	return p
}

// complete verifies command line arguments and loads data from the command environment
func (h *execHandle) complete(argsIn []string, conf config.ConfigProvider) {
	h.args = argsIn
	h.config = conf
	if h.environment == "" {
		h.environment = conf.GetStringQ(config.KubeEnvironmentName)
	}
	if h.service == "" {
		h.service = conf.GetStringQ(config.Service)
	}
}

// validate checks that the provided bash options are specified.
func (h *execHandle) validate() error {
	if len(strings.Trim(h.environment, " ")) == 0 {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.EnvironmentSpecifiedEmpty).String())
	}
	if len(strings.Trim(h.service, " ")) == 0 {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.ServiceSpecifiedEmpty).String())
	}
	return nil
}

// handle opens a bash console against a pod.
func (h *execHandle) handle(podsFinder pods.Finder, podsFilter pods.Filter) (suggestion string, err error) {
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionGetSettingsError, session.CurrentSession.SessionID), err
	}

	podsList, err := podsFinder.FindAll(user, apiKey, addr, h.environment)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionFindPodsFailed, session.CurrentSession.SessionID), err
	}

	pod := podsFilter.List(*podsList).ByService(h.service).ByStatus("Running").ByStatusReason("Running").First()
	if pod == nil {
		return fmt.Sprintf(msgs.SuggestionRunningPodNotFound, h.service, h.environment, config.AppName, "bash", session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, fmt.Sprintf(msgs.NoActivePodsFoundForSpecifiedServiceName, h.service)).String())
	}

	clientConfig := kubectlapi.GetNonInteractiveDeferredLoadingClientConfig(user, apiKey, addr, h.environment)
	kubeCmdExec := kubectlcmd.NewCmdExec(kubectlcmdutil.NewFactory(clientConfig), os.Stdin, os.Stdout, os.Stderr)
	kubeCmdExecOptions := &kubectlcmd.ExecOptions{
		StreamOptions: kubectlcmd.StreamOptions{
			In:  os.Stdin,
			Out: os.Stdout,
			Err: os.Stderr,
		},

		Executor: &kubectlcmd.DefaultRemoteExecutor{},
	}

	kubeCmdExecOptions.TTY = true
	kubeCmdExecOptions.Stdin = true
	kubeCmdExecOptions.PodName = pod.GetName()

	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {

		envTerm := os.Getenv("TERM")
		if envTerm == "" {
			envTerm = "xterm"
		}

		//ensure that the TERM environment variable is set
		//Work-around to be removed when kubernetes and docker fix the issue.
		//See docker/docker#26461 and kubernetes/kubernetes/issues/28280
		h.args = append([]string{"env", "TERM=" + envTerm}, h.args...)
	}

	kubeCmdUtilFactory := kubectlcmdutil.NewFactory(clientConfig)
	argsLenAtDash := kubeCmdExec.ArgsLenAtDash()
	err = kubeCmdExecOptions.Complete(kubeCmdUtilFactory, kubeCmdExec, h.args, argsLenAtDash)
	if err != nil {
		return "", errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, err.Error()).String())
	}
	err = kubeCmdExecOptions.Validate()
	if err != nil {
		return "", errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, err.Error()).String())
	}

	err = kubeCmdExecOptions.Run()
	if err != nil {
		cplogs.V(5).Infof("The pod may have been killed or moved to a different node. Error %s", err)
		cplogs.Flush()
		return fmt.Sprintf(msgs.SuggestionExecRunFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, err.Error()).String())
	}
	return "", nil
}

type interactiveModeHandler interface {
	Handle(flowID string, environment string) error
}

type interactiveModeH struct {
	config config.ConfigProvider
	api    cpapi.DataProvider
}

func newInteractiveModeH() *interactiveModeH {
	p := &interactiveModeH{}
	p.config = config.C
	p.api = cpapi.NewCpAPI()
	return p
}

func (h interactiveModeH) findTargetClusterAndApplyToConfig(flowID string, targetEnvironment string) (suggestion string, err error) {
	h.api.SetAPIKey(h.config.GetStringQ(config.ApiKey))

	environments, err := h.api.GetAPIEnvironments(flowID)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionGetApiEnvironmentsFailed, flowID, config.AppName, ExecCmdName, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cpapi.ErrorFailedToGetEnvironmentsList).String())
	}

	clusterIdentifier := ""
	for _, environment := range environments {
		if environment.Identifier == targetEnvironment {
			clusterIdentifier = environment.Cluster
		}
	}

	if clusterIdentifier == "" {
		return fmt.Sprintf(msgs.SuggestionEnvironmentListEmpty, targetEnvironment, flowID, config.AppName, ExecCmdName, session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.EnvironmentsNotFound).String())
	}

	//set the not persistent config information (**DO SAVE BY AND DO NOT CALL h.config.Save()** as we are in interactive mode)
	h.config.Set(config.CpKubeProxyEnabled, true)
	h.config.Set(config.FlowId, flowID)
	h.config.Set(config.ClusterIdentifier, clusterIdentifier)

	cplogs.V(5).Infof("interactive mode: flow set to %s", flowID)
	cplogs.V(5).Infof("interactive mode: cluster found and is set to %s", clusterIdentifier)
	cplogs.Flush()
	return "", nil
}
