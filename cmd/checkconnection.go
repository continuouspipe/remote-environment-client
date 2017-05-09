package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl"
)

//NewListPodsCmd return a cobra cmd alias of NewCheckConnectionCmd
func NewListPodsCmd() *cobra.Command {
	ck := NewCheckConnectionCmd()
	ck.Use = "pods"
	ck.Short = "Lists the pods in the remote environment (alias for checkconnection)"
	ck.Aliases = []string{"po"}
	return ck
}

//NewCheckConnectionCmd return a new cobra command that on run executes the CheckConnectionHandle
func NewCheckConnectionCmd() *cobra.Command {
	settings := config.C
	handler := &CheckConnectionHandle{}
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	command := &cobra.Command{
		Use:     "checkconnection",
		Aliases: []string{"ck"},
		Short:   "Check the connection to the remote environment",
		Long: `The checkconnection command can be used to check that the connection details
for the Kubernetes cluster are correct and lists any pods that can be found for the environment.
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
	command.PersistentFlags().StringVarP(&handler.Environment, config.KubeEnvironmentName, "e", environment, "The full remote environment name")

	return command
}

//CheckConnectionHandle holds the command handler dependencies
type CheckConnectionHandle struct {
	Command     *cobra.Command
	Environment string
	kubeCtlInit kubectlapi.KubeCtlInitializer
}

//Complete verifies command line arguments and loads data from the command environment
func (h *CheckConnectionHandle) Complete(cmd *cobra.Command, argsIn []string, setting *config.Config) error {
	h.Command = cmd
	var err error
	if h.Environment == "" {
		h.Environment, err = setting.GetString(config.KubeEnvironmentName)
		checkErr(err)
	}
	return nil
}

//Validate checks that the provided checkconnection options are specified.
func (h *CheckConnectionHandle) Validate() error {
	if len(strings.Trim(h.Environment, " ")) == 0 {
		//TODO: Reply with something more helpful
		return fmt.Errorf("the environment specified is invalid")
	}
	return nil
}

//Handle finds the pods and prints them
func (h *CheckConnectionHandle) Handle(args []string, podsFinder pods.Finder) error {
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return nil
	}

	fmt.Println("checking connection for environment " + h.Environment)

	podsList, err := podsFinder.FindAll(user, apiKey, addr, h.Environment)
	if err != nil {
		//TODO: Send error log to Sentry
		//TODO: Log err
		//TODO: Print user friendly error that explains what happened and what to do next
		return err
	}

	if len(podsList.Items) > 0 {
		printer := kubectl.NewHumanReadablePrinter(kubectl.PrintOptions{
			ColumnLabels:  []string{},
			Wide:          true,
			WithNamespace: true,
		})
		printer.EnsurePrintWithKind(podsList.Kind)
		color.Green("%d pods have been found:", len(podsList.Items))
		printer.PrintObj(podsList, os.Stdout)
	} else {
		color.Red("We could not find any pods on this environment")
	}

	return nil
}
