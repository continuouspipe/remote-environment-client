package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var bashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Open a bash session in the remote environment container",
	Long: `This will remotely connect to a bash session onto the default container specified
during setup but you can specify another container to connect to. `,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("bash called")
	},
}

func init() {
	RootCmd.AddCommand(bashCmd)

}
