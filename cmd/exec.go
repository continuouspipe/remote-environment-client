package cmd

import (
	"github.com/spf13/cobra"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/spf13/viper"
	"fmt"
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

	execCmd.PersistentFlags().StringP("pod", "p", "", "The pod to use")
}

type ExecHandle struct {
	Command *cobra.Command
}

func (h *ExecHandle) Handle(args []string) {
	validateConfig()

	context := viper.GetString("context")
	namespace := viper.GetString("namespace")
	pod := h.Command.PersistentFlags().Lookup("pod")

	fmt.Println(context)
	fmt.Println(namespace)
	fmt.Println(pod)

	kubectlapi.Exec(context, namespace, pod.Value.String(), "ls -all")
}
