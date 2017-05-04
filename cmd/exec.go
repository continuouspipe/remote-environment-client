package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/spf13/cobra"
	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	kubectlcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

var execExample = fmt.Sprintf(`
# execute -l -all on the web pod
%[1]s ex -- ls -all

# execute -l -all on the web pod overriding the project-key and remote-branch
%[1]s ex -e techup-dev-user -s web -- ls -all
`, config.AppName)

//NewExecCmd return a cobra command struct pointer which on Run, if required it prepares the config so we can reach the pod and
//then uses a command handler to execute the command specified in the arguments
func NewExecCmd() *cobra.Command {
	settings := config.C
	handler := newExecHandle()

	var interactive bool
	var flowID string

	bashcmd := &cobra.Command{
		Use:     "exec",
		Aliases: []string{"ex"},
		Short:   "Execute a command on a container",
		Long: `To execute a command on a container without first getting a bash session use
the exec command. The command and its arguments need to follow --`,
		Example: execExample,
		Run: func(cmd *cobra.Command, args []string) {
			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()
			local := exec.NewLocal()

			if interactive == true {
				cplogs.V(5).Infoln("exec in interactive mode")
				//make sure config has an api key and a cp user set
				initInteractiveH := NewInitInteractiveHandler(false)
				initInteractiveH.SetWriter(ioutil.Discard)
				err := initInteractiveH.Complete(args)
				checkErr(err)
				err = initInteractiveH.Validate()
				checkErr(err)
				err = initInteractiveH.Handle()
				checkErr(err)

				//alter the configuration so that we connect to the flow and environment specified by the user
				err = newInteractiveModeH().FindTargetClusterAndApplyToConfig(flowID, handler.environment)
				checkErr(err)
			}

			checkErr(handler.Complete(args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(podsFinder, podsFilter, local))
		},
	}

	defaultEnvironment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	defaultService, err := settings.GetString(config.Service)
	checkErr(err)
	defaultFlowID, err := settings.GetString(config.FlowId)
	checkErr(err)

	bashcmd.PersistentFlags().StringVarP(&handler.environment, config.KubeEnvironmentName, "e", defaultEnvironment, "The full remote environment name")
	bashcmd.PersistentFlags().StringVarP(&handler.service, config.Service, "s", defaultService, "The service to use (e.g.: web, mysql)")
	bashcmd.PersistentFlags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode requires the flags: --environment --service --flow to be specified")
	bashcmd.PersistentFlags().StringVarP(&flowID, config.FlowId, "f", defaultFlowID, "The flow to use")

	return bashcmd
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

// Complete verifies command line arguments and loads data from the command environment
func (h *execHandle) Complete(argsIn []string, conf config.ConfigProvider) error {
	h.args = argsIn
	h.config = conf
	var err error
	if h.environment == "" {
		h.environment, err = conf.GetString(config.KubeEnvironmentName)
		checkErr(err)
	}
	if h.service == "" {
		h.service, err = conf.GetString(config.Service)
		checkErr(err)
	}
	return nil
}

// Validate checks that the provided bash options are specified.
func (h *execHandle) Validate() error {
	if len(strings.Trim(h.environment, " ")) == 0 {
		return fmt.Errorf("the environment specified is invalid")
	}
	if len(strings.Trim(h.service, " ")) == 0 {
		return fmt.Errorf("the service specified is invalid")
	}
	return nil
}

// Handle opens a bash console against a pod.
func (h *execHandle) Handle(podsFinder pods.Finder, podsFilter pods.Filter, executor exec.Executor) error {
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return nil
	}

	podsList, err := podsFinder.FindAll(user, apiKey, addr, h.environment)
	if err != nil {
		return err
	}

	pod := podsFilter.List(*podsList).ByService(h.service).ByStatus("Running").ByStatusReason("Running").First()
	if pod == nil {
		return fmt.Errorf(fmt.Sprintf(msgs.NoActivePodsFoundForSpecifiedServiceName, h.service))
	}

	clientConfig := kubectlapi.GetNonInteractiveDeferredLoadingClientConfig(user, apiKey, addr, h.environment)
	kubeCmdExec := kubectlcmd.NewCmdExec(kubectlcmdutil.NewFactory(clientConfig), os.Stdin, os.Stdout, os.Stderr)
	kubeCmdExec.Flags().Set("pod", pod.GetName())
	kubeCmdExec.Flags().Set("stdin", "true")
	kubeCmdExec.Flags().Set("tty", "true")

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

	kubeCmdExec.Run(kubeCmdExec, h.args)
	return nil
}

type interactiveModeHandler interface {
	Handle(flowID string, environment string) error
}

type interactiveModeH struct {
	config config.ConfigProvider
	api    cpapi.CpApiProvider
}

func newInteractiveModeH() *interactiveModeH {
	p := &interactiveModeH{}
	p.config = config.C
	p.api = cpapi.NewCpApi()
	return p
}

func (h interactiveModeH) FindTargetClusterAndApplyToConfig(flowID string, targetEnvironment string) error {
	apiKey, err := h.config.GetString(config.ApiKey)
	if err != nil {
		return err
	}

	h.api.SetApiKey(apiKey)

	environments, el := h.api.GetApiEnvironments(flowID)
	if el != nil {
		return el
	}

	clusterIdentifier := ""
	for _, environment := range environments {
		if environment.Identifier == targetEnvironment {
			clusterIdentifier = environment.Cluster
		}
	}

	if clusterIdentifier == "" {
		return fmt.Errorf("The specified environment %s has not been found on the flow %s. Please check that you have inserted the correct flowID and environment at https://ui.continuouspipe.io/project/")
	}

	//set the not persistent config information (do not save the local config in interactive mode)
	h.config.Set(config.CpKubeProxyEnabled, true)
	h.config.Set(config.FlowId, flowID)
	cplogs.V(5).Infof("interactive mode: flow set to %s", flowID)
	h.config.Set(config.ClusterIdentifier, clusterIdentifier)
	cplogs.V(5).Infof("interactive mode: cluster found and is set to %s", clusterIdentifier)
	return nil
}
