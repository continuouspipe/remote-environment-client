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
		handler := &DestroyHandle{cmd}
		handler.Handle(args)
	},
}

func init() {
	RootCmd.AddCommand(destroyCmd)
}

type DestroyHandle struct {
	Command *cobra.Command
}

func (h *DestroyHandle) Handle(args []string) {
	validateConfig()
	fmt.Println("destroy called")
}