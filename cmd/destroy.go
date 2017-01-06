package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy the remote environment",
	Long: `The destroy command will delete the remote branch used for your remote
environment, ContinuousPipe will then remove the environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("destroy called")
	},
}

func init() {
	RootCmd.AddCommand(destroyCmd)
}
