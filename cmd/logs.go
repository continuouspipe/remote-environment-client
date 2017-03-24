package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/spf13/cobra"
	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	kubectlcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"os"
	"strconv"
	"strings"
	"time"
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

	command.PersistentFlags().DurationVar(&handler.since, "since", 0, "Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.")
	command.PersistentFlags().Int64Var(&handler.tail, "tail", -1, "Lines of recent log file to display. Defaults to -1, showing all log lines.")
	command.PersistentFlags().BoolVarP(&handler.follow, "follow", "f", false, "Specify if the logs should be streamed.")
	command.PersistentFlags().BoolVarP(&handler.previous, "previous", "p", false, "If true, print the logs for the previous instance of the container in a pod if it exists.")
	return command
}

type LogsCmdHandle struct {
	environment string
	service     string
	username    string
	apiKey      string
	conf        config.ConfigProvider
	kubeCtlInit kubectlapi.KubeCtlInitializer
	since       time.Duration
	tail        int64
	follow      bool
	previous    bool
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
	if h.username == "" {
		h.username, err = settings.GetString(config.Username)
		checkErr(err)
	}
	if h.apiKey == "" {
		h.apiKey, err = settings.GetString(config.ApiKey)
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
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return nil
	}

	allPods, err := podsFinder.FindAll(user, apiKey, addr, h.environment)
	if err != nil {
		return err
	}

	pod, err := podsFilter.ByService(allPods, h.service)
	if err != nil {
		return err
	}

	cplogs.V(5).Infof("getting container logs for environment %s, pod %s", h.environment, pod.GetName())
	cplogs.Flush()

	clientConfig := kubectlapi.GetNonInteractiveDeferredLoadingClientConfig(user, apiKey, addr, h.environment)
	kubeCmdLogs := kubectlcmd.NewCmdLogs(kubectlcmdutil.NewFactory(clientConfig), os.Stdout)

	kubeCmdLogs.Flags().Set("since", h.since.String())
	kubeCmdLogs.Flags().Set("tail", strconv.FormatInt(h.tail, 10))
	kubeCmdLogs.Flags().Set("follow", strconv.FormatBool(h.follow))
	kubeCmdLogs.Flags().Set("previous", strconv.FormatBool(h.previous))

	kubeCmdLogs.Run(kubeCmdLogs, []string{pod.GetName()})
	return nil
}
