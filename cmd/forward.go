package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/spf13/cobra"
	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	kubectlcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

var (
	portforwardExample = fmt.Sprintf(`
		# Listen on ports 5000 and 6000 locally, forwarding data to/from ports 5000 and 6000 in the pod
		%[1]s forward 5000 6000

		# Listen on port 8888 locally, forwarding to 5000 in the pod
		%[1]s forward 8888:5000

		# Listen on a random port locally, forwarding to 5000 in the pod
		%[1]s forward :5000

		# Listen on a random port locally, forwarding to 5000 in the pod
		%[1]s forward 0:5000

		# Overriding the project-key and remote-branch
		%[1]s forward -e techup-dev-user -s mysql 5000
		`, config.AppName)
)

func NewForwardCmd() *cobra.Command {
	settings := config.C
	handler := &ForwardHandle{}
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	command := &cobra.Command{
		Use:     "forward",
		Aliases: []string{"fo"},
		Short:   "Forward a port to a container",
		Long: `The forward command will set up port forwarding from the local environment
to a container on the remote environment that has a port exposed. This is useful for tasks
such as connecting to a database using a local client. You need to specify the container and
the port number to forward.`,
		Run: func(cmd *cobra.Command, args []string) {
			validateConfig()

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle())
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
func (h *ForwardHandle) Complete(cmd *cobra.Command, argsIn []string, settings *config.Config) error {
	h.Command = cmd
	h.ports = argsIn

	h.podsFinder = pods.NewKubePodsFind()
	h.podsFilter = pods.NewKubePodsFilter()

	var err error
	if h.Environment == "" {
		h.Environment, err = settings.GetString(config.KubeEnvironmentName)
		checkErr(err)
	}
	if h.Service == "" {
		h.Service, err = settings.GetString(config.Service)
		checkErr(err)
	}

	return nil
}

// Validate checks that the provided forward options are specified.
func (h *ForwardHandle) Validate() error {
	if len(h.ports) == 0 {
		return fmt.Errorf("at least 1 PORT is required for port-forward")
	}
	if len(strings.Trim(h.Environment, " ")) == 0 {
		return fmt.Errorf("the environment specified is invalid")
	}
	if len(strings.Trim(h.Service, " ")) == 0 {
		return fmt.Errorf("the service specified is invalid")
	}
	return nil
}

func (h *ForwardHandle) Handle() error {
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return nil
	}

	allPods, err := h.podsFinder.FindAll(user, apiKey, addr, h.Environment)
	if err != nil {
		return err
	}
	if err != nil {
		cplogs.V(5).Infof("pods not found for environment %s", h.Environment)
		return err
	}

	pod := h.podsFilter.List(*allPods).ByService(h.Service).ByStatus("Running").ByStatusReason("Running").First()
	if pod == nil {
		return fmt.Errorf(fmt.Sprintf(msgs.NoActivePodsFoundForSpecifiedServiceName, h.Service))
	}

	cplogs.V(5).Infof("setting up forwarding for target pod %s and ports %s", pod.GetName(), h.ports)
	cplogs.Flush()

	clientConfig := kubectlapi.GetNonInteractiveDeferredLoadingClientConfig(user, apiKey, addr, h.Environment)
	kubeCmdPortForward := kubectlcmd.NewCmdPortForward(kubectlcmdutil.NewFactory(clientConfig), os.Stdout, os.Stderr)
	kubeCmdPortForward.Run(kubeCmdPortForward, append([]string{pod.GetName()}, h.ports...))
	return nil
}
