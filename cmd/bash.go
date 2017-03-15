package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	kexec "github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
)

var bashInteractiveFullExample = fmt.Sprintf(`%s bash --interactive ([-i]) --environment ([-e]) php-example-cpdev-foo --service ([-s]) web --flow-id ([-f]) 1268cc54-b265-11e6-b835-0c360641bb54`, config.AppName)

var bashExample = fmt.Sprintf(`
Default Mode:

	# bash into the default remote container
	%[1]s bash
	%[1]s ba

	# bash into a different environment and service within the same flow
	%[1]s bash -e techup-dev-user -s web

Interactive Mode:

	# bash into a different environment and service in a different flow
	%[2]s`, config.AppName, bashInteractiveFullExample)

func NewBashCmd() *cobra.Command {
	settings := config.C
	handler := &bashHandle{}
	handler.api = cpapi.NewCpApi()
	handler.config = settings
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()

	bashcmd := &cobra.Command{
		Use:     "bash",
		Aliases: []string{"ba"},
		Short:   "Open a bash session in the remote environment container",
		Long: `This will remotely connect to a bash session onto the default container specified
during setup but you can specify another container to connect to. `,
		Run: func(cmd *cobra.Command, args []string) {
			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()
			local := exec.NewLocal()

			checkErr(handler.Complete(args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(args, podsFinder, podsFilter, local))
		},
		Example: bashExample,
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

type bashHandle struct {
	config      *config.Config
	environment string
	service     string
	flowId      string
	interactive bool
	api         cpapi.CpApiProvider
	kubeCtlInit kubectlapi.KubeCtlInitializer
}

// Complete verifies command line arguments and loads data from the command environment
func (h *bashHandle) Complete(argsIn []string, conf *config.Config) error {
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
func (h *bashHandle) Validate() error {
	if len(strings.Trim(h.environment, " ")) == 0 {
		return fmt.Errorf("the environment specified is invalid")
	}
	if len(strings.Trim(h.service, " ")) == 0 {
		return fmt.Errorf("the service specified is invalid")
	}
	return nil
}

// Handle opens a bash console against a pod.
func (h *bashHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter, executor exec.Executor) error {

	if h.interactive {
		cplogs.V(5).Infoln("bash in interactive mode")
		initInteractiveHandler := NewInitInteractiveHandler(false)
		initInteractiveHandler.writer = ioutil.Discard
		err := initInteractiveHandler.Complete(args)
		if err != nil {
			return err
		}
		err = initInteractiveHandler.Validate()
		if err != nil {
			return err
		}
		err = initInteractiveHandler.Handle()
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

	err := h.kubeCtlInit.Init(h.environment)
	if err != nil {
		return err
	}

	podsList, err := podsFinder.FindAll(h.environment, h.environment)
	if err != nil {
		return err
	}

	pod, err := podsFilter.ByService(podsList, h.service)
	if err != nil {
		return err
	}

	kscmd := kexec.KSCommand{}
	kscmd.KubeConfigKey = h.environment
	kscmd.Environment = h.environment
	kscmd.Pod = pod.GetName()
	kscmd.Stdin = os.Stdin
	kscmd.Stdout = os.Stdout
	kscmd.Stderr = os.Stderr

	return executor.StartProcess(kscmd, "/bin/bash")
}
