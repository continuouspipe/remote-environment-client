package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
		settings := config.NewApplicationSettings()
		validator := config.NewMandatoryChecker()
		validateConfig(validator, settings)

		handler := &ExecHandle{cmd}
		podsFinder := pods.NewKubePodsFind()
		podsFilter := pods.NewKubePodsFilter()
		local := exec.NewLocal()

		res := handler.Handle(args, settings, podsFinder, podsFilter, local)
		fmt.Println(res)
	},
}

func init() {
	RootCmd.AddCommand(execCmd)
}

type ExecHandle struct {
	Command *cobra.Command
}

func (h *ExecHandle) Handle(args []string, settings config.Reader, podsFinder pods.Finder, podsFilter pods.Filter, spawn exec.Spawner) string {
	kubeConfigKey := settings.GetString(config.KubeConfigKey)
	environment := settings.GetString(config.Environment)
	service := settings.GetString(config.Service)

	allPods, err := podsFinder.FindAll(kubeConfigKey, environment)
	checkErr(err)

	pod, err := podsFilter.ByService(allPods, service)
	checkErr(err)

	return spawn.CommandExec(kubeConfigKey, environment, pod.GetName(), args...)
}
