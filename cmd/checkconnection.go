package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
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
		context := viper.GetString("key")
		environment := viper.GetString("environment")
		fmt.Println("checking connection for environment " + environment)

		config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{CurrentContext: context}).ClientConfig()
		if err != nil {
			panic(err.Error())
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		pods, err := clientset.Core().Pods(environment).List(v1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	},
}

func init() {
	RootCmd.AddCommand(checkconnectionCmd)

	checkconnectionCmd.PersistentFlags().StringP("environment", "e", "", "The environment to use")
}
