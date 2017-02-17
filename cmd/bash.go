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

	projectKey, err := settings.GetString(config.ProjectKey)
	checkErr(err)
	remoteBranch, err := settings.GetString(config.RemoteBranch)
	checkErr(err)
	service, err := settings.GetString(config.Service)
	checkErr(err)

	bashcmd.PersistentFlags().StringVarP(&handler.ProjectKey, config.ProjectKey, "p", projectKey, "Continuous Pipe project key")
	bashcmd.PersistentFlags().StringVarP(&handler.RemoteBranch, config.RemoteBranch, "r", remoteBranch, "Name of the Git branch you are using for your remote environment")
	bashcmd.PersistentFlags().StringVarP(&handler.Service, config.Service, "s", service, "The service to use (e.g.: web, mysql)")

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
func (h *BashHandle) Complete(cmd *cobra.Command, argsIn []string, conf *config.Config) error {
	h.Command = cmd

	var err error

	h.kubeConfigKey, err = conf.GetString(config.KubeConfigKey)
	checkErr(err)
	if h.ProjectKey == "" {
		h.ProjectKey, err = conf.GetString(config.ProjectKey)
		checkErr(err)
	}
	if h.RemoteBranch == "" {
		h.RemoteBranch, err = conf.GetString(config.RemoteBranch)
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

	kscmd := kexec.KSCommand{}
	kscmd.KubeConfigKey = h.kubeConfigKey
	kscmd.Environment = environment
	kscmd.Pod = pod.GetName()
	kscmd.Stdin = os.Stdin
	kscmd.Stdout = os.Stdout
	kscmd.Stderr = os.Stderr

	return executor.StartProcess(kscmd, "/bin/bash")
}
