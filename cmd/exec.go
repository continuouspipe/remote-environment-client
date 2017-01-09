package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a command on a container",
	Long: `To execute a command on a container without first getting a bash session use
the exec command. The command and its arguments need to follow --`,
	Run: func(cmd *cobra.Command, args []string) {
		handler := &ExecHandle{cmd}
		handler.Handle(args)
	},
}

func init() {
	RootCmd.AddCommand(execCmd)
}

type ExecHandle struct {
	Command *cobra.Command
}

func (h *ExecHandle) Handle(args []string) {
	validateConfig()
	fmt.Println("exec called")
}