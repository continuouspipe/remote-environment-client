package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/config"
	"fmt"
)

var (
	portforward_example = fmt.Sprintf(`
		# Listen on ports 5000 and 6000 locally, forwarding data to/from ports 5000 and 6000 in the pod
		%s forward 5000 6000

		# Listen on port 8888 locally, forwarding to 5000 in the pod
		%s forward 8888:5000

		# Listen on a random port locally, forwarding to 5000 in the pod
		%s forward :5000

		# Listen on a random port locally, forwarding to 5000 in the pod
		%s forward 0:5000`, config.AppName, config.AppName, config.AppName, config.AppName)
)

var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Forward a port to a container",
	Long: `The forward command will set up port forwarding from the local environment
to a container on the remote environment that has a port exposed. This is useful for tasks 
such as connecting to a database using a local client. You need to specify the container and 
the port number to forward.`,
	Run: func(cmd *cobra.Command, args []string) {
		handler := &ForwardHandle{cmd}
		if err := handler.Validate(args); err != nil {
			checkErr(err)
		}
		handler.Handle(args)
	},
	Example: portforward_example,
}

func init() {
	RootCmd.AddCommand(forwardCmd)
}

type ForwardHandle struct {
	Command *cobra.Command
}

func (h *ForwardHandle) Handle(args []string) {
	validateConfig()

	kubeConfigKey := viper.GetString(config.KubeConfigKey)
	environment := viper.GetString(config.Environment)
	service := viper.GetString(config.Service)

	pod, err := kubectlapi.FindPodByService(kubeConfigKey, environment, service)
	checkErr(err)

	ports := args[0]
	kubectlapi.Forward(pod.GetName(), ports)
}

func (h *ForwardHandle) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("at least 1 PORT is required for port-forward")
	}
	return nil
}
