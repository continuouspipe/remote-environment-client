package cmd

import (
	"fmt"

	envconfig "github.com/continuouspipe/remote-environment-client/config"
	"github.com/spf13/cobra"
)

var bashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Open a bash session in the remote environment container",
	Long: `This will remotely connect to a bash session onto the default container specified
during setup but you can specify another container to connect to. `,
	Run: func(cmd *cobra.Command, args []string) {
		RunCommand(cmd, args)
	},
}

func RunCommand(cmd *cobra.Command, args []string) {
	i, missing := envconfig.Validate()
	if i > 0 {
		fmt.Printf("The remote settings file is missing or the require parameters are missing (%v), please run the setup command.", missing)
		return
	}
	fmt.Println("bash called")
}

func init() {
	RootCmd.AddCommand(bashCmd)

}
