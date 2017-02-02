package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/spf13/cobra"
	"strings"
	"os"
	"github.com/continuouspipe/remote-environment-client/sync/monitor"
	"time"
	"io"
)

func NewWatchCmd() *cobra.Command {
	settings := config.NewApplicationSettings()
	handler := &WatchHandle{}
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
			validator := config.NewMandatoryChecker()
			validateConfig(validator, settings)

			exclusion := monitor.NewExclusion()
			exclusion.LoadCustomExclusionsFromFile()

			//if the custom file/folder exclusions are empty write the default one on disk
			if len(exclusion.CustomExclusions) == 0 {
				res, err := exclusion.WriteDefaultExclusionsToFile()
				checkErr(err)
				if res == true {
					fmt.Printf("\n%s was missing or empty and has been created with the default ignore settings.\n", sync.SyncExcluded)
				}
			}

			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()

			handler.Stdout = os.Stdout

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(dirMonitor, podsFinder, podsFilter))
		},
	}
	command.PersistentFlags().StringVarP(&handler.ProjectKey, config.ProjectKey, "p", settings.GetString(config.ProjectKey), "Continuous Pipe project key")
	command.PersistentFlags().StringVarP(&handler.RemoteBranch, config.RemoteBranch, "r", settings.GetString(config.RemoteBranch), "Name of the Git branch you are using for your remote environment")
	command.PersistentFlags().StringVarP(&handler.Service, config.Service, "s", settings.GetString(config.Service), "The service to use (e.g.: web, mysql)")
	command.PersistentFlags().Int64VarP(&handler.Latency, "latency", "l", 500, "Sync latency / speed in milli-seconds")
	command.PersistentFlags().IntVarP(&handler.IndividualFileSyncThreshold, "individual-file-sync-threshold", "t", 10, "Above this threshold the watch command will sync any file or folder that is different compared to the local one")
	return command
}

type WatchHandle struct {
	Command                     *cobra.Command
	ProjectKey                  string
	RemoteBranch                string
	Service                     string
	kubeConfigKey               string
	Latency                     int64
	Stdout                      io.Writer
	IndividualFileSyncThreshold int
}

// Complete verifies command line arguments and loads data from the command environment
func (h *WatchHandle) Complete(cmd *cobra.Command, argsIn []string, settingsReader config.Reader) error {
	h.Command = cmd

	h.kubeConfigKey = settingsReader.GetString(config.KubeConfigKey)

	if h.ProjectKey == "" {
		h.ProjectKey = settingsReader.GetString(config.ProjectKey)
	}
	if h.RemoteBranch == "" {
		h.RemoteBranch = settingsReader.GetString(config.RemoteBranch)
	}
	if h.Service == "" {
		h.Service = settingsReader.GetString(config.Service)
	}

	return nil
}

// Validate checks that the provided watch options are specified.
func (h *WatchHandle) Validate() error {
	if len(strings.Trim(h.ProjectKey, " ")) == 0 {
		return fmt.Errorf("the project key specified is invalid")
	}
	if len(strings.Trim(h.RemoteBranch, " ")) == 0 {
		return fmt.Errorf("the remote branch specified is invalid")
	}
	if len(strings.Trim(h.Service, " ")) == 0 {
		return fmt.Errorf("the service specified is invalid")
	}
	if h.Latency <= 100 {
		return fmt.Errorf("please specify a latency of at least 100 milli-seconds")
	}
	return nil
}

func (h *WatchHandle) Handle(dirMonitor monitor.DirectoryMonitor, podsFinder pods.Finder, podsFilter pods.Filter) error {
	environment := config.GetEnvironment(h.ProjectKey, h.RemoteBranch)

	allPods, err := podsFinder.FindAll(h.kubeConfigKey, environment)
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

	observer := sync.GetDirectoryEventSyncAll()
	observer.KubeConfigKey = h.kubeConfigKey
	observer.Environment = environment
	observer.Pod = *pod
	observer.IndividualFileSyncThreshold = h.IndividualFileSyncThreshold
	dirMonitor.SetLatency(time.Duration(h.Latency))

	fmt.Fprintf(h.Stdout, "\nDestination Pod: %s\n", pod.GetName())

	return dirMonitor.AnyEventCall(cwd, observer)
}
