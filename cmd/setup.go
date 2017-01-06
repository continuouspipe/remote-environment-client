package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the remote environment client and settings",
	Long: `This will ask a series of questions to get the details for the project set up. 
	    
Your answers will be stored in a .cp-remote-env-settings file in the project root. You 
will probably want to add this to your .gitignore file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("setup called")
	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}
