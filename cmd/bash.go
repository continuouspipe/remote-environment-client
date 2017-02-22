package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	kexec "github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func NewBashCmd() *cobra.Command {
	settings := config.C
	handler := &BashHandle{}

	bashcmd := &cobra.Command{
		Use:     "bash",
		Aliases: []string{"ba"},
		Short:   "Open a bash session in the remote environment container",
		Long: `This will remotely connect to a bash session onto the default container specified
during setup but you can specify another container to connect to. `,
		Run: func(cmd *cobra.Command, args []string) {
			validateConfig()

			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()
			local := exec.NewLocal()

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(args, podsFinder, podsFilter, local))
		},
		Example: fmt.Sprintf("%s bash", config.AppName),
	}

	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	service, err := settings.GetString(config.Service)
	checkErr(err)

	bashcmd.PersistentFlags().StringVarP(&handler.Environment, config.KubeEnvironmentName, "r", environment, "The full remote environment name: project-key-git-branch")
	bashcmd.PersistentFlags().StringVarP(&handler.Service, config.Service, "s", service, "The service to use (e.g.: web, mysql)")

	return bashcmd
}

type BashHandle struct {
	Command     *cobra.Command
	Environment string
	Service     string
}

// Complete verifies command line arguments and loads data from the command environment
func (h *BashHandle) Complete(cmd *cobra.Command, argsIn []string, conf *config.Config) error {
	h.Command = cmd

	var err error
	if h.Environment == "" {
		h.Environment, err = conf.GetString(config.KubeEnvironmentName)
		checkErr(err)
	}
	if h.Service == "" {
		h.Service, err = conf.GetString(config.Service)
		checkErr(err)
	}

	return nil
}

// Validate checks that the provided bash options are specified.
func (h *BashHandle) Validate() error {
	if len(strings.Trim(h.Environment, " ")) == 0 {
		return fmt.Errorf("the environment specified is invalid")
	}
	if len(strings.Trim(h.Service, " ")) == 0 {
		return fmt.Errorf("the service specified is invalid")
	}
	return nil
}

// Handle opens a bash console against a pod.
func (h *BashHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter, executor exec.Executor) error {
	podsList, err := podsFinder.FindAll(h.Environment, h.Environment)
	if err != nil {
		return err
	}

	pod, err := podsFilter.ByService(podsList, h.Service)
	if err != nil {
		return err
	}

	kscmd := kexec.KSCommand{}
	kscmd.KubeConfigKey = h.Environment
	kscmd.Environment = h.Environment
	kscmd.Pod = pod.GetName()
	kscmd.Stdin = os.Stdin
	kscmd.Stdout = os.Stdout
	kscmd.Stderr = os.Stderr

	return executor.StartProcess(kscmd, "/bin/bash")
}
