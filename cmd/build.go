package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Create/Update the remote environment",
	Long: `The build command will push changes the branch you have checked out locally to your remote 
environment branch. ContinuousPipe will then build the environment. You can use the 
https://ui.continuouspipe.io/ to see when the environment has finished building and to 
find its IP address.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("build called")
	},
}

func init() {
	RootCmd.AddCommand(buildCmd)
}
