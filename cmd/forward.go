package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/spf13/cobra"
	"strings"
)

var (
	portforwardExample = fmt.Sprintf(`
		# Listen on ports 5000 and 6000 locally, forwarding data to/from ports 5000 and 6000 in the pod
		%s forward 5000 6000

		# Listen on port 8888 locally, forwarding to 5000 in the pod
		%s forward 8888:5000

		# Listen on a random port locally, forwarding to 5000 in the pod
		%s forward :5000

		# Listen on a random port locally, forwarding to 5000 in the pod
		%s forward 0:5000

		# Overriding the project-key and remote-branch
		%s forward -p techup -r dev-user -s mysql 5000
		`, config.AppName, config.AppName, config.AppName, config.AppName, config.AppName)
)

func NewForwardCmd() *cobra.Command {
	settings := config.NewApplicationSettings()
	handler := &ForwardHandle{}
	command := &cobra.Command{
		Use:     "forward",
		Aliases: []string{"fo"},
		Short:   "Forward a port to a container",
		Long: `The forward command will set up port forwarding from the local environment
to a container on the remote environment that has a port exposed. This is useful for tasks
such as connecting to a database using a local client. You need to specify the container and
the port number to forward.`,
		Run: func(cmd *cobra.Command, args []string) {
			validator := config.NewMandatoryChecker()
			validateConfig(validator, settings)

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle())
		},
		Example: portforwardExample,
	}
	command.PersistentFlags().StringVarP(&handler.ProjectKey, config.ProjectKey, "p", settings.GetString(config.ProjectKey), "Continuous Pipe project key")
	command.PersistentFlags().StringVarP(&handler.RemoteBranch, config.RemoteBranch, "r", settings.GetString(config.RemoteBranch), "Name of the Git branch you are using for your remote environment")
	command.PersistentFlags().StringVarP(&handler.Service, config.Service, "s", settings.GetString(config.Service), "The service to use (e.g.: web, mysql)")
	return command
}

type ForwardHandle struct {
	Command       *cobra.Command
	ports         string
	podsFinder    pods.Finder
	podsFilter    pods.Filter
	ProjectKey    string
	RemoteBranch  string
	Service       string
	kubeConfigKey string
}

// Complete verifies command line arguments and loads data from the command environment
func (h *ForwardHandle) Complete(cmd *cobra.Command, argsIn []string, settingsReader config.Reader) error {
	h.Command = cmd

	if len(argsIn) > 0 {
		h.ports = argsIn[0]
	}

	h.podsFinder = pods.NewKubePodsFind()
	h.podsFilter = pods.NewKubePodsFilter()
	h.kubeConfigKey = settingsReader.GetString(config.KubeConfigKey)

	if h.ProjectKey == "" {
		h.ProjectKey = settingsReader.GetString(config.ProjectKey)
	}
	if h.RemoteBranch == "" {
		h.RemoteBranch = settingsReader.GetString(config.RemoteBranch)
	}
	if h.Service == "" {
		h.Service = settingsReader.GetString(config.Service)
	}

	return nil
}

// Validate checks that the provided forward options are specified.
func (h *ForwardHandle) Validate() error {
	if len(h.ports) == 0 {
		return fmt.Errorf("at least 1 PORT is required for port-forward")
	}
	if len(strings.Trim(h.ProjectKey, " ")) == 0 {
		return fmt.Errorf("the project key specified is invalid")
	}
	if len(strings.Trim(h.RemoteBranch, " ")) == 0 {
		return fmt.Errorf("the remote branch specified is invalid")
	}
	if len(strings.Trim(h.Service, " ")) == 0 {
		return fmt.Errorf("the service specified is invalid")
	}
	return nil
}

func (h *ForwardHandle) Handle() error {
	environment := config.GetEnvironment(h.ProjectKey, h.RemoteBranch)

	allPods, err := h.podsFinder.FindAll(h.kubeConfigKey, environment)
	if err != nil {
		cplogs.V(5).Infof("pods not found for project key %s and remote branch %s", h.ProjectKey, h.RemoteBranch)
		return err
	}

	pod, err := h.podsFilter.ByService(allPods, h.Service)
	if err != nil {
		cplogs.V(5).Infof("pods not found for service %s", h.Service)
		return err
	}

	cplogs.V(5).Infof("setting up forwarding for target pod %s and ports %s", pod.GetName(), h.ports)
	cplogs.Flush()
	return kubectlapi.Forward(h.kubeConfigKey, environment, pod.GetName(), h.ports)
}
