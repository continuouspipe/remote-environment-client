package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/config"
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
}

type ExecHandle struct {
	Command *cobra.Command
}

func (h *ExecHandle) Handle(args []string) {
	validateConfig()

	kubeConfigKey := viper.GetString(config.KubeConfigKey)
	environment := viper.GetString(config.Environment)
	service := viper.GetString(config.Service)

	pod, err := kubectlapi.FindPodByService(kubeConfigKey, environment, service)
	checkErr(err)

	res := kubectlapi.Exec(kubeConfigKey, environment, pod.GetName(), args...)
	fmt.Println(res)
}
