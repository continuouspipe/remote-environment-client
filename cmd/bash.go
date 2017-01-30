package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/spf13/cobra"
	"strings"
)

func NewBashCmd() *cobra.Command {
	settings := config.NewApplicationSettings()
	handler := &BashHandle{}

	bashcmd := &cobra.Command{
		Use:     "bash",
		Aliases: []string{"ba"},
		Short:   "Open a bash session in the remote environment container",
		Long: `This will remotely connect to a bash session onto the default container specified
during setup but you can specify another container to connect to. `,
		Run: func(cmd *cobra.Command, args []string) {
			validator := config.NewMandatoryChecker()
			validateConfig(validator, settings)

			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()
			local := exec.NewLocal()

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(args, podsFinder, podsFilter, local))
		},
		Example: fmt.Sprintf("%s bash", config.AppName),
	}

	bashcmd.PersistentFlags().StringVarP(&handler.ProjectKey, config.ProjectKey, "p", settings.GetString(config.ProjectKey), "Continuous Pipe project key")
	bashcmd.PersistentFlags().StringVarP(&handler.RemoteBranch, config.RemoteBranch, "r", settings.GetString(config.RemoteBranch), "Name of the Git branch you are using for your remote environment")
	bashcmd.PersistentFlags().StringVarP(&handler.Service, config.Service, "s", settings.GetString(config.Service), "The service to use (e.g.: web, mysql)")

	return bashcmd
}

type BashHandle struct {
	Command       *cobra.Command
	ProjectKey    string
	RemoteBranch  string
	Service       string
	kubeConfigKey string
}

// Complete verifies command line arguments and loads data from the command environment
func (h *BashHandle) Complete(cmd *cobra.Command, argsIn []string, settingsReader config.Reader) error {
	h.Command = cmd

	h.kubeConfigKey = settingsReader.GetString(config.KubeConfigKey)

	if h.ProjectKey == "" {
		h.ProjectKey = settingsReader.GetString(config.ProjectKey)
	}
	if h.RemoteBranch == "" {
		h.RemoteBranch = settingsReader.GetString(config.RemoteBranch)
	}
	if h.Service == "" {
		h.Service = settingsReader.GetString(config.Service)
	}

	return nil
}

// Validate checks that the provided bash options are specified.
func (h *BashHandle) Validate() error {
	if len(strings.Trim(h.ProjectKey, " ")) == 0 {
		return fmt.Errorf("the project key specified is invalid")
	}
	if len(strings.Trim(h.RemoteBranch, " ")) == 0 {
		return fmt.Errorf("the remote branch specified is invalid")
	}
	if len(strings.Trim(h.Service, " ")) == 0 {
		return fmt.Errorf("the service specified is invalid")
	}
	return nil
}

// Handle opens a bash console against a pod.
func (h *BashHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter, executor exec.Executor) error {
	environment := config.GetEnvironment(h.ProjectKey, h.RemoteBranch)

	podsList, err := podsFinder.FindAll(h.kubeConfigKey, environment)
	if err != nil {
		return err
	}

	pod, err := podsFilter.ByService(podsList, h.Service)
	if err != nil {
		return err
	}

	return executor.StartProcess(h.kubeConfigKey, environment, pod.GetName(), "/bin/bash")
}
