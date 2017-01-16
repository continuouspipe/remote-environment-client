package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/config"
)

var bashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Open a bash session in the remote environment container",
	Long: `This will remotely connect to a bash session onto the default container specified
during setup but you can specify another container to connect to. `,
	Run: func(cmd *cobra.Command, args []string) {
		handler := &BashHandle{cmd}
		settings := config.NewApplicationSettings()
		podsFinder := pods.NewKubePodsFind()
		podsFilter := pods.NewKubePodsFilter()
		local := exec.NewLocal()
		handler.Handle(args, settings, podsFinder, podsFilter, local)
	},
	Example: fmt.Sprintf("%s bash", config.AppName),
}

type BashHandle struct {
	Command *cobra.Command
}

func init() {
	RootCmd.AddCommand(bashCmd)
}

func (h *BashHandle) Handle(args []string, settingsReader config.Reader, podsFinder pods.Finder, podsFilter pods.Filter, executor exec.Executor) {
	validateConfig()

	kubeConfigKey := settingsReader.GetString(config.KubeConfigKey)
	environment := settingsReader.GetString(config.Environment)
	service := settingsReader.GetString(config.Service)

	podsList, err := podsFinder.FindAll(kubeConfigKey, environment)
	checkErr(err)

	pod, err := podsFilter.ByService(podsList, service)
	checkErr(err)

	executor.SysCallExec(kubeConfigKey, environment, pod.GetName(), "/bin/bash")
}
