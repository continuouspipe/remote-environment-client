package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a command on a container",
	Long: `To execute a command on a container without first getting a bash session use
the exec command. The command and its arguments need to follow --`,
	Run: func(cmd *cobra.Command, args []string) {
		handler := &ExecHandle{cmd}
		podsFinder := pods.NewKubePodsFind()
		podsFilter := pods.NewKubePodsFilter()
		local := exec.NewLocal()
		handler.Handle(args, podsFinder, podsFilter, local)
	},
}

func init() {
	RootCmd.AddCommand(execCmd)
}

type ExecHandle struct {
	Command *cobra.Command
}

func (h *ExecHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter, spawn exec.Spawner) {
	validateConfig()

	kubeConfigKey := viper.GetString(config.KubeConfigKey)
	environment := viper.GetString(config.Environment)
	service := viper.GetString(config.Service)

	allPods, err := podsFinder.FindAll(kubeConfigKey, environment)
	checkErr(err)

	pod, err := podsFilter.ByService(allPods, service)
	checkErr(err)

	res := spawn.CommandExec(kubeConfigKey, environment, pod.GetName(), args...)
	fmt.Println(res)
}
