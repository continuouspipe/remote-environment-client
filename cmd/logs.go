package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/spf13/cobra"
	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	kubectlcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"os"
	"strings"
)

var (
	logs_example = fmt.Sprintf(`
		# Return snapshot logs from the pod that matches the default service
		%[1]s logs

		# Return snapshot logs from pod mysql with only one container
		%[1]s logs mysql

		# Return snapshot of previous terminated ruby container logs from pod mysql
		%[1]s logs -p mysql

		# Begin streaming the logs of the ruby container in pod mysql
		%[1]s logs -f mysql

		# Display only the most recent 20 lines of output in pod mysql
		%[1]s logs --tail=20 mysql

		# Show all logs from pod mysql written in the last hour
		%[1]s logs --since=1h mysql`, config.AppName)
)

func NewLogsCmd() *cobra.Command {
	settings := config.C
	handler := &LogsCmdHandle{}
	handler.kubeLogsOptions = kubectlcmd.NewCmdLogs(kubectlcmdutil.NewFactory(nil), os.Stdout)
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	command := &cobra.Command{
		Use:     "logs",
		Aliases: []string{"lo"},
		Short:   "Print the logs for a pod",
		Long:    `Print the logs for a pod`,
		Run: func(cmd *cobra.Command, args []string) {
			validateConfig()

			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(args, podsFinder, podsFilter))
		},
	}
	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	service, err := settings.GetString(config.Service)
	checkErr(err)

	//used to find the targed pod
	command.PersistentFlags().StringVarP(&handler.environment, config.KubeEnvironmentName, "e", environment, "The full remote environment name: project-key-git-branch")
	command.PersistentFlags().StringVarP(&handler.service, config.Service, "s", service, "The service to use (e.g.: web, mysql)")

	//will be passed to the kubectl logs command
	since, err := handler.kubeLogsOptions.Flags().GetDuration("since")
	checkErr(err)
	tail, err := handler.kubeLogsOptions.Flags().GetInt64("tail")
	checkErr(err)
	follow, err := handler.kubeLogsOptions.Flags().GetBool("follow")
	checkErr(err)
	previous, err := handler.kubeLogsOptions.Flags().GetBool("previous")
	checkErr(err)

	command.PersistentFlags().DurationVar(&since, "since", 0, "Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.")
	command.PersistentFlags().Int64Var(&tail, "tail", -1, "Lines of recent log file to display. Defaults to -1, showing all log lines.")
	command.PersistentFlags().BoolVarP(&follow, "follow", "f", false, "Specify if the logs should be streamed.")
	command.PersistentFlags().BoolVarP(&previous, "previous", "p", false, "If true, print the logs for the previous instance of the container in a pod if it exists.")
	return command
}

type LogsCmdHandle struct {
	environment     string
	service         string
	kubeCtlInit     kubectlapi.KubeCtlInitializer
	kubeLogsOptions *cobra.Command
}

// Complete verifies command line arguments and loads data from the command environment
func (h *LogsCmdHandle) Complete(cmd *cobra.Command, argsIn []string, settings *config.Config) error {
	var err error
	if h.environment == "" {
		h.environment, err = settings.GetString(config.KubeEnvironmentName)
		checkErr(err)
	}
	if h.service == "" {
		h.service, err = settings.GetString(config.Service)
		checkErr(err)
	}
	return nil
}

// Validate checks that the provided exec options are specified.
func (h *LogsCmdHandle) Validate() error {
	if len(strings.Trim(h.environment, " ")) == 0 {
		return fmt.Errorf("the environment specified is invalid")
	}
	if len(strings.Trim(h.service, " ")) == 0 {
		return fmt.Errorf("the service specified is invalid")
	}
	return nil
}

// Handle set the flags in the kubeclt logs handle command and executes it
func (h *LogsCmdHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter) error {
	//re-init kubectl in case the kube settings have been modified
	err := h.kubeCtlInit.Init(h.environment)
	if err != nil {
		return nil
	}

	allPods, err := podsFinder.FindAll(h.environment, h.environment)
	checkErr(err)

	pod, err := podsFilter.ByService(allPods, h.service)
	checkErr(err)

	h.kubeLogsOptions.Run(h.kubeLogsOptions, []string{pod.GetName()})

	return nil
}
