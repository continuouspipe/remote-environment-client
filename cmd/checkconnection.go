package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"strings"
)

func NewCheckConnectionCmd() *cobra.Command {
	settings := config.C
	handler := &CheckConnectionHandle{}
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	command := &cobra.Command{
		Use:     "checkconnection",
		Aliases: []string{"ck"},
		Short:   "Check the connection to the remote environment",
		Long: `The checkconnection command can be used to check that the connection details
for the Kubernetes cluster are correct and that if they are pods can be found for the environment.
It can be used with the environment option to check another environment`,
		Run: func(cmd *cobra.Command, args []string) {
			validateConfig()

			podsFinder := pods.NewKubePodsFind()
			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(args, podsFinder))
		},
	}

	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	command.PersistentFlags().StringVarP(&handler.Environment, config.KubeEnvironmentName, "e", environment, "The full remote environment name: project-key-git-branch")

	return command
}

type CheckConnectionHandle struct {
	Command       *cobra.Command
	Environment   string
	kubeConfigKey string
	kubeCtlInit   kubectlapi.KubeCtlInitializer
}

// Complete verifies command line arguments and loads data from the command environment
func (h *CheckConnectionHandle) Complete(cmd *cobra.Command, argsIn []string, setting *config.Config) error {
	h.Command = cmd
	var err error
	if h.Environment == "" {
		h.Environment, err = setting.GetString(config.KubeEnvironmentName)
		checkErr(err)
	}
	return nil
}

// Validate checks that the provided checkconnection options are specified.
func (h *CheckConnectionHandle) Validate() error {
	if len(strings.Trim(h.Environment, " ")) == 0 {
		return fmt.Errorf("the environment specified is invalid")
	}
	return nil
}

// Finds the pods and prints them
func (h *CheckConnectionHandle) Handle(args []string, podsFinder pods.Finder) error {
	//re-init kubectl in case the kube settings have been modified
	h.kubeCtlInit.Init()

	fmt.Println("checking connection for environment " + h.Environment)

	countPods, err := fetchNumberOfPods(h.kubeConfigKey, h.Environment, podsFinder)
	if err != nil {
		return err
	}
	color.Green("Connected succesfully and found %d pods for the environment\n", countPods)
	return nil
}

func fetchNumberOfPods(kubeConfigKey string, environment string, podsFinder pods.Finder) (int, error) {
	foundPods, err := podsFinder.FindAll(kubeConfigKey, environment)
	if err != nil {
		return 0, err
	}
	if len(foundPods.Items) == 0 {
		return 0, fmt.Errorf("connected to the cluster but no pods were found for the environment, has the environment been successfully built?")
	}
	return len(foundPods.Items), nil
}
