package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/benchmark"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/continuouspipe/remote-environment-client/sync/monitor"
	"github.com/spf13/cobra"
	"path/filepath"
	"strings"
)

var pushSyncExample = `
# push files and folders to the remote pod
%[1]s %[2]s

# push files and folders to the remote pod specifying the environment
%[1]s %[2]s -e techup-dev-user -s web
`

func NewSyncCmd() *cobra.Command {
	pu := NewPushCmd()
	pu.Use = "sync"
	pu.Short = "Sync local changes to the remote filesystem (alias for push)"
	pu.Aliases = []string{"sy"}
	pu.Example = fmt.Sprintf(pushSyncExample, config.AppName, "sync")
	return pu
}

func NewPushCmd() *cobra.Command {
	settings := config.C
	handler := &PushHandle{}
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()

	command := &cobra.Command{
		Use:     "push",
		Aliases: []string{"pu"},
		Short:   "Push local changes to the remote filesystem",
		Example: fmt.Sprintf(pushSyncExample, config.AppName, "push"),
		Long: `The push command will copy changes from the local to the remote filesystem.
Note that this will delete any files/folders in the remote container that are not present locally.`,
		Run: func(cmd *cobra.Command, args []string) {
			validateConfig()

			exclusion := monitor.NewExclusion()
			_, err := exclusion.WriteDefaultExclusionsToFile()
			checkErr(err)

			fmt.Println("Push in progress")

			benchmrk := benchmark.NewCmdBenchmark()
			benchmrk.Start("push")

			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()
			syncer := sync.GetSyncer()

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(args, podsFinder, podsFilter, syncer))

			_, err = benchmrk.StopAndLog()
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
	command.PersistentFlags().StringVarP(&handler.File, "file", "f", "", "Allows to specify a file that needs to be pushed to the pod")
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
func (h *PushHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter, syncer sync.Syncer) error {
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return nil
	}

	allPods, err := podsFinder.FindAll(user, apiKey, addr, h.Environment)
	if err != nil {
		return err
	}

	pod, err := podsFilter.ByService(allPods, h.Service)
	if err != nil {
		return err
	}

	//set individual file threshold to 1 as for now we only allow the user to specify 1 file to be pushed
	syncer.SetIndividualFileSyncThreshold(1)

	syncer.SetEnvironment(h.Environment)
	syncer.SetKubeConfigKey(h.Environment)
	syncer.SetPod(pod.GetName())
	syncer.SetRemoteProjectPath(h.RemoteProjectPath)

	var paths []string
	if h.File != "" {
		absFilePath, err := filepath.Abs(h.File)
		if err != nil {
			return err
		}
		paths = append(paths, absFilePath)
	}

	return syncer.Sync(paths)
}
