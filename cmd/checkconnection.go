package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCheckConnectionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "checkconnection",
		Short: "Check the connection to the remote environment",
		Long: `The checkconnection command can be used to check that the connection details
for the Kubernetes cluster are correct and that if they are pods can be found for the environment.
It can be used with the environment option to check another environment`,
		Run: func(cmd *cobra.Command, args []string) {
			settings := config.NewApplicationSettings()
			validator := config.NewMandatoryChecker()
			validateConfig(validator, settings)

			handler := &CheckConnectionHandle{cmd}
			podsFinder := pods.NewKubePodsFind()
			handler.Handle(args, podsFinder)
		},
	}
}

type CheckConnectionHandle struct {
	Command *cobra.Command
}

func (h *CheckConnectionHandle) Handle(args []string, podsFinder pods.Finder) {
	viper.BindPFlag("environment", h.Command.PersistentFlags().Lookup("environment"))
	kubeConfigKey := viper.GetString("kubernetes-config-key")
	environment := viper.GetString("environment")
	fmt.Println("checking connection for environment " + environment)
	color.Green("Connected succesfully and found %d pods for the environment\n", fetchNumberOfPods(kubeConfigKey, environment, podsFinder))
}

func fetchNumberOfPods(kubeConfigKey string, environment string, podsFinder pods.Finder) int {
	foundPods, err := podsFinder.FindAll(kubeConfigKey, environment)
	checkErr(err)

	if len(foundPods.Items) == 0 {
		exitWithMessage("connected to the cluster but no pods were found for the environment, has the environment been successfully built?")
	}
	return len(foundPods.Items)
}
