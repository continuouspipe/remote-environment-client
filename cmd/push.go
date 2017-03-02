package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/benchmark"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/spf13/cobra"
	"strings"
)

var pushExample = `
# push files and folders to the remote pod
cp-push pu

# push files and folders to the remote pod specifying the environment
cp-push pu -e techup-dev-user -s web
`

func NewPushCmd() *cobra.Command {
	settings := config.C
	handler := &PushHandle{}
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()

	command := &cobra.Command{
		Use:     "push",
		Aliases: []string{"pu"},
		Short:   "Push local changes to the remote filesystem",
		Example: pushExample,
		Long: `The push command will copy changes from the local to the remote filesystem.
		Note that this will delete any files/folders in the remote container that are not present locally.`,
		Run: func(cmd *cobra.Command, args []string) {
			validateConfig()

			fmt.Println("Push in progress")

			benchmark := benchmark.NewCmdBenchmark()
			benchmark.Start("push")

			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()
			pusher := sync.GetPusher()

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(args, podsFinder, podsFilter, pusher))

			_, err := benchmark.StopAndLog()
			checkErr(err)
			fmt.Printf("Push complete, the files and folders that has been sent can be found in the logs %s\n", cplogs.GetLogInfoFile())
			cplogs.Flush()
		},
	}

	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	service, err := settings.GetString(config.Service)
	checkErr(err)

	command.PersistentFlags().StringVarP(&handler.Environment, config.KubeEnvironmentName, "e", environment, "The full remote environment name: project-key-git-branch")
	command.PersistentFlags().StringVarP(&handler.Service, config.Service, "s", service, "The service to use (e.g.: web, mysql)")
	command.PersistentFlags().StringVarP(&handler.File, "file", "f", "", "Allows to specify a file that needs to be pushed from the pod")
	command.PersistentFlags().StringVarP(&handler.RemoteProjectPath, "remote-project-path", "a", "/app/", "Specify the absolute path to your project folder, by default set to /app/")
	return command
}

type PushHandle struct {
	Command           *cobra.Command
	Environment       string
	Service           string
	File              string
	RemoteProjectPath string
	kubeCtlInit       kubectlapi.KubeCtlInitializer
}

// Complete verifies command line arguments and loads data from the command environment
func (h *PushHandle) Complete(cmd *cobra.Command, argsIn []string, settings *config.Config) error {
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
	if strings.HasSuffix(h.RemoteProjectPath, "/") == false {
		h.RemoteProjectPath = h.RemoteProjectPath + "/"
	}
	return nil
}

// Validate checks that the provided push options are specified.
func (h *PushHandle) Validate() error {
	if len(strings.Trim(h.Environment, " ")) == 0 {
		return fmt.Errorf("the environment specified is invalid")
	}
	if len(strings.Trim(h.Service, " ")) == 0 {
		return fmt.Errorf("the service specified is invalid")
	}
	if strings.HasPrefix(h.RemoteProjectPath, "/") == false {
		return fmt.Errorf("please specify an absolute path for your --remote-project-path")
	}
	return nil
}

// Copies all the files and folders from the current directory into the remote container
func (h *PushHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter, pusher sync.Pusher) error {
	//re-init kubectl in case the kube settings have been modified
	err := h.kubeCtlInit.Init(h.Environment)
	if err != nil {
		return err
	}

	allPods, err := podsFinder.FindAll(h.Environment, h.Environment)
	if err != nil {
		return err
	}

	pod, err := podsFilter.ByService(allPods, h.Service)
	if err != nil {
		return err
	}

	pusher.SetEnvironment(h.Environment)
	pusher.SetKubeConfigKey(h.Environment)
	pusher.SetPod(pod.GetName())
	pusher.SetRemoteProjectPath(h.RemoteProjectPath)
	return pusher.Push(h.File)
}
