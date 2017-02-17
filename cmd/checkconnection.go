package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"strings"
)

func NewCheckConnectionCmd() *cobra.Command {
	settings := config.C
	handler := &CheckConnectionHandle{}
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

	projectKey, err := settings.GetString(config.ProjectKey)
	checkErr(err)
	remoteBranch, err := settings.GetString(config.RemoteBranch)
	checkErr(err)

	command.PersistentFlags().StringVarP(&handler.ProjectKey, config.ProjectKey, "p", projectKey, "Continuous Pipe project key")
	command.PersistentFlags().StringVarP(&handler.RemoteBranch, config.RemoteBranch, "r", remoteBranch, "Name of the Git branch you are using for your remote environment")

	return command
}

type CheckConnectionHandle struct {
	Command       *cobra.Command
	ProjectKey    string
	RemoteBranch  string
	kubeConfigKey string
}

// Complete verifies command line arguments and loads data from the command environment
func (h *CheckConnectionHandle) Complete(cmd *cobra.Command, argsIn []string, setting *config.Config) error {
	h.Command = cmd

	var err error
	h.kubeConfigKey, err = setting.GetString(config.KubeConfigKey)
	checkErr(err)

	if h.ProjectKey == "" {
		h.ProjectKey, err = setting.GetString(config.ProjectKey)
		checkErr(err)
	}
	if h.RemoteBranch == "" {
		h.RemoteBranch, err = setting.GetString(config.RemoteBranch)
		checkErr(err)
	}

	return nil
}

// Validate checks that the provided checkconnection options are specified.
func (h *CheckConnectionHandle) Validate() error {
	if len(strings.Trim(h.ProjectKey, " ")) == 0 {
		return fmt.Errorf("the project key specified is invalid")
	}
	if len(strings.Trim(h.RemoteBranch, " ")) == 0 {
		return fmt.Errorf("the remote branch specified is invalid")
	}
	return nil
}

// Finds the pods and prints them
func (h *CheckConnectionHandle) Handle(args []string, podsFinder pods.Finder) error {
	environment := config.GetEnvironment(h.ProjectKey, h.RemoteBranch)

	fmt.Println("checking connection for environment " + environment)

	countPods, err := fetchNumberOfPods(h.kubeConfigKey, environment, podsFinder)
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
