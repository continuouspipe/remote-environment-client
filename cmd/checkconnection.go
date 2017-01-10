package cmd

import (
	"fmt"

	"github.com/continuouspipe/remote-environment-client/kubeapi"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var checkconnectionCmd = &cobra.Command{
	Use:   "checkconnection",
	Short: "Check the connection to the remote environment",
	Long: `The checkconnection command can be used to check that the connection details
for the Kubernetes cluster are correct and that if they are pods can be found for the environment. 
It can be used with the environment option to check another environment`,
	Run: func(cmd *cobra.Command, args []string) {
		handler := &CheckConnectionHandle{cmd}
		handler.Handle(args)
	},
}

func init() {
	RootCmd.AddCommand(checkconnectionCmd)

	checkconnectionCmd.PersistentFlags().StringP("environment", "e", "", "The environment to use")
}

type CheckConnectionHandle struct {
	Command *cobra.Command
}

func (h *CheckConnectionHandle) Handle(args []string) {
	validateConfig()

	viper.BindPFlag("environment", h.Command.PersistentFlags().Lookup("environment"))
	context := viper.GetString("kubernetes-config-key")
	environment := viper.GetString("environment")
	fmt.Println("checking connection for environment " + environment)
	color.Green("Connected succesfully and found %d pods for the environment\n", fetchNumberOfPods(context, environment))
}

func fetchNumberOfPods(context string, environment string) int {
	pods, err := kubeapi.FetchPods(context, environment)
	checkErr(err)

	if len(pods.Items) == 0 {
		exitWithMessage("connected to the cluster but no pods were found for the environment, has the environment been successfully built?")
	}
	return len(pods.Items)
}
