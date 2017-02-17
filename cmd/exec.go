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
	settings := config.NewApplicationSettings()
	handler := &ExecHandle{}
	command := &cobra.Command{
		Use:     "exec",
		Aliases: []string{"ex"},
		Short:   "Execute a command on a container",
		Long: `To execute a command on a container without first getting a bash session use
the exec command. The command and its arguments need to follow --`,
		Example: execExample,
		Run: func(cmd *cobra.Command, args []string) {
			validator := config.NewMandatoryChecker()
			validateConfig(validator, settings)

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
	command.PersistentFlags().StringVarP(&handler.ProjectKey, config.ProjectKey, "p", settings.GetString(config.ProjectKey), "Continuous Pipe project key")
	command.PersistentFlags().StringVarP(&handler.RemoteBranch, config.RemoteBranch, "r", settings.GetString(config.RemoteBranch), "Name of the Git branch you are using for your remote environment")
	command.PersistentFlags().StringVarP(&handler.Service, config.Service, "s", settings.GetString(config.Service), "The service to use (e.g.: web, mysql)")
	return command
}

type ExecHandle struct {
	Command       *cobra.Command
	ProjectKey    string
	RemoteBranch  string
	Service       string
	kubeConfigKey string
}

// Complete verifies command line arguments and loads data from the command environment
func (h *ExecHandle) Complete(cmd *cobra.Command, argsIn []string, settingsReader config.Reader) error {
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

// Validate checks that the provided exec options are specified.
func (h *ExecHandle) Validate() error {
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

// Handle executes a command inside a pod
func (h *ExecHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter, spawn exec.Spawner) (string, error) {
	environment := config.GetEnvironment(h.ProjectKey, h.RemoteBranch)

	allPods, err := podsFinder.FindAll(h.kubeConfigKey, environment)
	checkErr(err)

	pod, err := podsFilter.ByService(allPods, h.Service)
	checkErr(err)

	kscmd := kexec.KSCommand{}
	kscmd.KubeConfigKey = h.kubeConfigKey
	kscmd.Environment = environment
	kscmd.Pod = pod.GetName()
	kscmd.Stdin = os.Stdin
	kscmd.Stdout = os.Stdout
	kscmd.Stderr = os.Stderr

	return spawn.CommandExec(kscmd, args...)
}
