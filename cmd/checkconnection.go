package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var checkconnectionCmd = &cobra.Command{
	Use:   "checkconnection",
	Short: "Check the connection to the remote environment",
	Long: `The checkconnection command can be used to check that the connection details
for the Kubernetes cluster are correct and that if they are pods can be found for the environment. 
It can be used with the environment option to check another environment`,
	Run: func(cmd *cobra.Command, args []string) {

		viper.BindPFlag("environment", cmd.PersistentFlags().Lookup("environment"))
		context := viper.GetString("kubernetes-config-key")
		environment := viper.GetString("environment")
		fmt.Println("checking connection for environment " + environment)
		color.Green("Connected succesfully and found %d pods for the environment\n", fetchNumberOfPods(createClient(readConfig(context)), environment))
	},
}

func init() {
	RootCmd.AddCommand(checkconnectionCmd)

	checkconnectionCmd.PersistentFlags().StringP("environment", "e", "", "The environment to use")
}

func checkErr(err error) {
	if err != nil {
		exitWithMessage(err.Error())
	}
}

func exitWithMessage(message string) {
	color.Set(color.FgRed)
	fmt.Println("ERROR: " + message)
	color.Unset()
	os.Exit(1)
}

func readConfig(context string) *rest.Config {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: context}).ClientConfig()
	checkErr(err)
	return config
}

func createClient(config *rest.Config) *kubernetes.Clientset {
	client, err := kubernetes.NewForConfig(config)
	checkErr(err)
	return client
}

func fetchNumberOfPods(client *kubernetes.Clientset, environment string) int {
	pods, err := client.Core().Pods(environment).List(v1.ListOptions{})
	checkErr(err)

	if len(pods.Items) == 0 {
		exitWithMessage("connected to the cluster but no pods were found for the environment, has the environment been successfully built?")
	}
	return len(pods.Items)
}
