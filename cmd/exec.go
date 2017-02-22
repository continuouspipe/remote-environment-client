package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	kexec "github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/spf13/cobra"
)

var execExample = `
# execute -l -all on the web pod
cp-remote ex -- ls -all

# execute -l -all on the web pod overriding the project-key and remote-branch
cp-remote ex -p techup -r dev-user -s web -- ls -all
`

func NewExecCmd() *cobra.Command {
	settings := config.C
	handler := &ExecHandle{}
	command := &cobra.Command{
		Use:     "exec",
		Aliases: []string{"ex"},
		Short:   "Execute a command on a container",
		Long: `To execute a command on a container without first getting a bash session use
the exec command. The command and its arguments need to follow --`,
		Example: execExample,
		Run: func(cmd *cobra.Command, args []string) {
			validateConfig()

			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()
			local := exec.NewLocal()

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			res, err := handler.Handle(args, podsFinder, podsFilter, local)
			checkErr(err)
			fmt.Println(res)
		},
	}
	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	service, err := settings.GetString(config.Service)
	checkErr(err)

	command.PersistentFlags().StringVarP(&handler.Environment, config.KubeEnvironmentName, "r", environment, "The full remote environment name: project-key-git-branch")
	command.PersistentFlags().StringVarP(&handler.Service, config.Service, "s", service, "The service to use (e.g.: web, mysql)")
	return command
}

type ExecHandle struct {
	Command     *cobra.Command
	Environment string
	Service     string
}

// Complete verifies command line arguments and loads data from the command environment
func (h *ExecHandle) Complete(cmd *cobra.Command, argsIn []string, settings *config.Config) error {
	h.Command = cmd
	var err error
	if h.Environment == "" {
		h.Environment, err = settings.GetString(config.KubeEnvironmentName)
		checkErr(err)
	}
	if h.Service == "" {
		h.Service, err = settings.GetString(config.Service)
		checkErr(err)
	}

	return nil
}

// Validate checks that the provided exec options are specified.
func (h *ExecHandle) Validate() error {
	if len(strings.Trim(h.Environment, " ")) == 0 {
		return fmt.Errorf("the environment specified is invalid")
	}
	if len(strings.Trim(h.Service, " ")) == 0 {
		return fmt.Errorf("the service specified is invalid")
	}
	return nil
}

// Handle executes a command inside a pod
func (h *ExecHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter, spawn exec.Spawner) (string, error) {
	allPods, err := podsFinder.FindAll(h.Environment, h.Environment)
	checkErr(err)

	pod, err := podsFilter.ByService(allPods, h.Service)
	checkErr(err)

	kscmd := kexec.KSCommand{}
	kscmd.KubeConfigKey = h.Environment
	kscmd.Environment = h.Environment
	kscmd.Pod = pod.GetName()
	kscmd.Stdin = os.Stdin
	kscmd.Stdout = os.Stdout
	kscmd.Stderr = os.Stderr

	return spawn.CommandExec(kscmd, args...)
}
