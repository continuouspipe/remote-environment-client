package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	kexec "github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var execExample = fmt.Sprintf(`
# execute -l -all on the web pod
%[1]s ex -- ls -all

# execute -l -all on the web pod overriding the project-key and remote-branch
%[1]s ex -e techup-dev-user -s web -- ls -all
`, config.AppName)

func NewExecCmd() *cobra.Command {
	settings := config.C
	handler := NewExecHandle()
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

			checkErr(handler.Complete(args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(podsFinder, podsFilter, local))
		},
	}

	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	service, err := settings.GetString(config.Service)
	checkErr(err)
	flowId, err := settings.GetString(config.FlowId)
	checkErr(err)

	bashcmd.PersistentFlags().StringVarP(&handler.environment, config.KubeEnvironmentName, "e", environment, "The full remote environment name: project-key-git-branch")
	bashcmd.PersistentFlags().StringVarP(&handler.service, config.Service, "s", service, "The service to use (e.g.: web, mysql)")
	bashcmd.PersistentFlags().BoolVarP(&handler.interactive, "interactive", "i", false, "Interactive mode requires the flags: --environment --service --flow to be specified")
	bashcmd.PersistentFlags().StringVarP(&handler.flowId, config.FlowId, "f", flowId, "The flow to use")

	return bashcmd
}

type execHandle struct {
	args             []string
	config           config.ConfigProvider
	environment      string
	service          string
	flowId           string
	interactive      bool
	api              cpapi.CpApiProvider
	kubeCtlInit      kubectlapi.KubeCtlInitializer
	initInteractiveH InitStrategy
}

func NewExecHandle() *execHandle {
	p := &execHandle{}
	p.api = cpapi.NewCpApi()
	p.config = config.C
	p.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	p.initInteractiveH = NewInitInteractiveHandler(false)
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
	if h.interactive {
		if len(strings.Trim(h.flowId, " ")) == 0 {
			return fmt.Errorf("the flowId specified is invalid")
		}
	}
	return nil
}

// Handle opens a bash console against a pod.
func (h *execHandle) Handle(podsFinder pods.Finder, podsFilter pods.Filter, executor exec.Executor) error {

	if h.interactive {
		cplogs.V(5).Infoln("bash in interactive mode")
		h.initInteractiveH.SetWriter(ioutil.Discard)
		err := h.initInteractiveH.Complete(h.args)
		if err != nil {
			return err
		}
		err = h.initInteractiveH.Validate()
		if err != nil {
			return err
		}
		err = h.initInteractiveH.Handle()
		if err != nil {
			return err
		}

		apiKey, err := h.config.GetString(config.ApiKey)
		if err != nil {
			return err
		}

		h.api.SetApiKey(apiKey)

		environments, err := h.api.GetApiEnvironments(h.flowId)
		if err != nil {
			return err
		}

		clusterIdentifier := ""
		for _, environment := range environments {
			if environment.Identifier == h.environment {
				clusterIdentifier = environment.Cluster
			}
		}

		if clusterIdentifier == "" {
			return fmt.Errorf("The specified environment %s has not been found on the flow %s. Please check that you have inserted the correct flowId and environment at https://ui.continuouspipe.io/project/")
		}

		//set the not persistent config information (do not save the local config in interactive mode)
		h.config.Set(config.CpKubeProxyEnabled, true)
		h.config.Set(config.FlowId, h.flowId)
		cplogs.V(5).Infof("interactive mode: flow set to %s", h.flowId)
		h.config.Set(config.ClusterIdentifier, clusterIdentifier)
		cplogs.V(5).Infof("interactive mode: cluster found and is set to %s", clusterIdentifier)
	}

	//TODO: Remove this when we get rid of the dependency on ~/.kube/config and call directly the KubeExecCmd without spawning
	err := h.kubeCtlInit.Init(h.environment)
	if err != nil {
		return err
	}

	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return nil
	}

	podsList, err := podsFinder.FindAll(user, apiKey, addr, h.environment)
	if err != nil {
		return err
	}

	pod, err := podsFilter.ByService(podsList, h.service)
	if err != nil {
		return err
	}

	//TODO: Change to call directly the KubeCtl NewCmdExec()
	kscmd := kexec.KSCommand{}
	kscmd.KubeConfigKey = h.environment
	kscmd.Environment = h.environment
	kscmd.Pod = pod.GetName()
	kscmd.Stdin = os.Stdin
	kscmd.Stdout = os.Stdout
	kscmd.Stderr = os.Stderr

	return executor.StartProcess(kscmd, h.args...)
}
