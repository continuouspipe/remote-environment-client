package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/config"
)

var bashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Open a bash session in the remote environment container",
	Long: `This will remotely connect to a bash session onto the default container specified
during setup but you can specify another container to connect to. `,
	Run: func(cmd *cobra.Command, args []string) {
		handler := &BashHandle{cmd}
		handler.Handle(args)
	},
}

type BashHandle struct {
	Command *cobra.Command
}

func init() {
	RootCmd.AddCommand(bashCmd)
}

func (h *BashHandle) Handle(args []string) {
	validateConfig()

	kubeConfigKey := viper.GetString(config.KubeConfigKey)
	environment := viper.GetString(config.Environment)
	service := viper.GetString(config.Service)

	pod, err := kubectlapi.FindPodByService(kubeConfigKey, environment, service)
	checkErr(err)

	kubectlapi.SysCallExec(kubeConfigKey, environment, pod.GetName(), "/bin/bash")
}
