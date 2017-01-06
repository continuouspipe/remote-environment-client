package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Forward a port to a container",
	Long: `The forward command will set up port forwarding from the local environment
to a container on the remote environment that has a port exposed. This is useful for tasks 
such as connecting to a database using a local client. You need to specify the container and 
the port number to forward.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("forward called")
	},
}

func init() {
	RootCmd.AddCommand(forwardCmd)
}
