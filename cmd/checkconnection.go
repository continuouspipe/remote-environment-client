package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var checkconnectionCmd = &cobra.Command{
	Use:   "checkconnection",
	Short: "Check the connection to the remote environment",
	Long: `The checkconnection command can be used to check that the connection details
for the Kubernetes cluster are correct and that if they are pods can be found for the environment. 
It can be used with the environment option to check another environment`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("checkconnection called")
	},
}

func init() {
	RootCmd.AddCommand(checkconnectionCmd)
}
