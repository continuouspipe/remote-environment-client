package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/continuouspipe/remote-environment-client/sync/monitor"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
	"time"
)

func NewWatchCmd() *cobra.Command {
	settings := config.C
	handler := &WatchHandle{}
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	command := &cobra.Command{
		Use:     "watch",
		Aliases: []string{"wa"},
		Short:   "Watch local changes and synchronize with the remote environment",
		Long: `The watch command will sync changes you make locally to a container that's part
of the remote environment. This will use the default container specified during
setup but you can specify another container to sync with.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Watching for changes. Quit anytime with Ctrl-C.")

			dirMonitor := monitor.GetOsDirectoryMonitor()
			validateConfig()

			exclusion := monitor.NewExclusion()
			exclusion.LoadCustomExclusionsFromFile()

			//if the custom file/folder exclusions are empty write the default one on disk
			if len(exclusion.CustomExclusions) == 0 {
				res, err := exclusion.WriteDefaultExclusionsToFile()
				checkErr(err)
				if res == true {
					fmt.Printf("\n%s was missing or empty and has been created with the default ignore settings.\n", monitor.CustomExclusionsFile)
				}
			}

			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()

			handler.Stdout = os.Stdout
			handler.syncer = sync.GetSyncer()

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(dirMonitor, podsFinder, podsFilter))
		},
	}

	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	service, err := settings.GetString(config.Service)
	checkErr(err)

	command.PersistentFlags().StringVarP(&handler.Environment, config.KubeEnvironmentName, "e", environment, "The full remote environment name: project-key-git-branch")
	command.PersistentFlags().StringVarP(&handler.Service, config.Service, "s", service, "The service to use (e.g.: web, mysql)")
	command.PersistentFlags().Int64VarP(&handler.Latency, "latency", "l", 500, "Sync latency / speed in milli-seconds")
	command.PersistentFlags().IntVarP(&handler.IndividualFileSyncThreshold, "individual-file-sync-threshold", "t", 10, "Above this threshold the watch command will sync any file or folder that is different compared to the local one")
	command.PersistentFlags().StringVarP(&handler.RemoteProjectPath, "remote-project-path", "a", "/app/", "Specify the absolute path to your project folder, by default set to /app/")
	return command
}

type WatchHandle struct {
	Command                     *cobra.Command
	Environment                 string
	Service                     string
	Latency                     int64
	Stdout                      io.Writer
	IndividualFileSyncThreshold int
	RemoteProjectPath           string
	syncer                      sync.Syncer
	kubeCtlInit                 kubectlapi.KubeCtlInitializer
}

// Complete verifies command line arguments and loads data from the command environment
func (h *WatchHandle) Complete(cmd *cobra.Command, argsIn []string, settings *config.Config) error {
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

// Validate checks that the provided watch options are specified.
func (h *WatchHandle) Validate() error {
	if len(strings.Trim(h.Environment, " ")) == 0 {
		return fmt.Errorf("the environment specified is invalid")
	}
	if len(strings.Trim(h.Service, " ")) == 0 {
		return fmt.Errorf("the service specified is invalid")
	}
	if h.Latency <= 100 {
		return fmt.Errorf("please specify a latency of at least 100 milli-seconds")
	}
	if strings.HasPrefix(h.RemoteProjectPath, "/") == false {
		return fmt.Errorf("please specify an absolute path for your --remote-project-path")
	}
	return nil
}

func (h *WatchHandle) Handle(dirMonitor monitor.DirectoryMonitor, podsFinder pods.Finder, podsFilter pods.Filter) error {
	//re-init kubectl in case the kube settings have been modified
	h.kubeCtlInit.Init()

	allPods, err := podsFinder.FindAll(h.Environment, h.Environment)
	if err != nil {
		return err
	}

	pod, err := podsFilter.ByService(allPods, h.Service)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	h.syncer.SetKubeConfigKey(h.Environment)
	h.syncer.SetEnvironment(h.Environment)
	h.syncer.SetPod(pod.GetName())
	h.syncer.SetIndividualFileSyncThreshold(h.IndividualFileSyncThreshold)
	h.syncer.SetRemoteProjectPath(h.RemoteProjectPath)
	dirMonitor.SetLatency(time.Duration(h.Latency))

	fmt.Fprintf(h.Stdout, "\nDestination Pod: %s\n", pod.GetName())

	observer := sync.GetSyncOnEventObserver(h.syncer)

	return dirMonitor.AnyEventCall(cwd, observer)
}
