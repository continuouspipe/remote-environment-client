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
	settings := config.NewApplicationSettings()
	handler := &CheckConnectionHandle{}
	command := &cobra.Command{
		Use:     "checkconnection",
		Aliases: []string{"ck"},
		Short:   "Check the connection to the remote environment",
		Long: `The checkconnection command can be used to check that the connection details
for the Kubernetes cluster are correct and that if they are pods can be found for the environment.
It can be used with the environment option to check another environment`,
		Run: func(cmd *cobra.Command, args []string) {
			validator := config.NewMandatoryChecker()
			validateConfig(validator, settings)

			podsFinder := pods.NewKubePodsFind()
			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(args, podsFinder))
		},
	}

	command.PersistentFlags().StringVarP(&handler.ProjectKey, config.ProjectKey, "p", settings.GetString(config.ProjectKey), "Continuous Pipe project key")
	command.PersistentFlags().StringVarP(&handler.RemoteBranch, config.RemoteBranch, "r", settings.GetString(config.RemoteBranch), "Name of the Git branch you are using for your remote environment")

	return command
}

type CheckConnectionHandle struct {
	Command       *cobra.Command
	ProjectKey    string
	RemoteBranch  string
	kubeConfigKey string
}

// Complete verifies command line arguments and loads data from the command environment
func (h *CheckConnectionHandle) Complete(cmd *cobra.Command, argsIn []string, settingsReader config.Reader) error {
	h.Command = cmd

	h.kubeConfigKey = settingsReader.GetString(config.KubeConfigKey)

	if h.ProjectKey == "" {
		h.ProjectKey = settingsReader.GetString(config.ProjectKey)
	}
	if h.RemoteBranch == "" {
		h.RemoteBranch = settingsReader.GetString(config.RemoteBranch)
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
