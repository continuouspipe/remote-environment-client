package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/continuouspipe/remote-environment-client/sync/monitor"
	"github.com/continuouspipe/remote-environment-client/sync/options"
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
	handler.api = cpapi.NewCpApi()
	handler.config = settings
	handler.writer = os.Stdout

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
			_, err := exclusion.WriteDefaultExclusionsToFile()
			checkErr(err)

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

	command.PersistentFlags().StringVarP(&handler.options.environment, config.KubeEnvironmentName, "e", environment, "The full remote environment name: project-key-git-branch")
	command.PersistentFlags().StringVarP(&handler.options.service, config.Service, "s", service, "The service to use (e.g.: web, mysql)")
	command.PersistentFlags().Int64VarP(&handler.options.latency, "latency", "l", 500, "Sync latency / speed in milli-seconds")
	command.PersistentFlags().IntVarP(&handler.options.individualFileSyncThreshold, "individual-file-sync-threshold", "t", 10, "Above this threshold the watch command will sync any file or folder that is different compared to the local one")
	command.PersistentFlags().StringVarP(&handler.options.remoteProjectPath, "remote-project-path", "a", "/app/", "Specify the absolute path to your project folder, by default set to /app/")
	command.PersistentFlags().BoolVar(&handler.options.dryRun, "dry-run", false, "Show what would have been transferred")
	command.PersistentFlags().BoolVar(&handler.options.rsyncVerbose, "rsync-verbose", false, "Allows to use rsync in verbose mode and debug issues with exclusions")
	command.PersistentFlags().BoolVar(&handler.options.delete, "delete", false, "Delete extraneous files from destination directories")
	return command
}

type WatchHandle struct {
	Command     *cobra.Command
	Stdout      io.Writer
	syncer      sync.Syncer
	kubeCtlInit kubectlapi.KubeCtlInitializer
	api         cpapi.CpApiProvider
	config      config.ConfigProvider
	writer      io.Writer
	options     watchCmdOptions
}

type watchCmdOptions struct {
	environment, service, remoteProjectPath string
	latency                                 int64
	individualFileSyncThreshold             int
	rsyncVerbose, dryRun, delete            bool
}

// Complete verifies command line arguments and loads data from the command environment
func (h *WatchHandle) Complete(cmd *cobra.Command, argsIn []string, settings *config.Config) error {
	h.Command = cmd

	var err error
	if h.options.environment == "" {
		h.options.environment, err = settings.GetString(config.KubeEnvironmentName)
		checkErr(err)
	}
	if h.options.service == "" {
		h.options.service, err = settings.GetString(config.Service)
		checkErr(err)
	}
	if strings.HasSuffix(h.options.remoteProjectPath, "/") == false {
		h.options.remoteProjectPath = h.options.remoteProjectPath + "/"
	}
	return nil
}

// Validate checks that the provided watch options are specified.
func (h *WatchHandle) Validate() error {
	if len(strings.Trim(h.options.environment, " ")) == 0 {
		return fmt.Errorf("the environment specified is invalid")
	}
	if len(strings.Trim(h.options.service, " ")) == 0 {
		return fmt.Errorf("the service specified is invalid")
	}
	if h.options.latency <= 100 {
		return fmt.Errorf("please specify a latency of at least 100 milli-seconds")
	}
	if strings.HasPrefix(h.options.remoteProjectPath, "/") == false {
		return fmt.Errorf("please specify an absolute path for your --remote-project-path")
	}
	return nil
}

func (h *WatchHandle) Handle(dirMonitor monitor.DirectoryMonitor, podsFinder pods.Finder, podsFilter pods.Filter) error {
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return nil
	}

	allPods, err := podsFinder.FindAll(user, apiKey, addr, h.options.environment)
	if err != nil {
		return err
	}

	pod, err := podsFilter.ByService(allPods, h.options.service)
	if err != nil {
		return err
	}

	remoteEnvId, err := h.config.GetString(config.RemoteEnvironmentId)
	if err != nil {
		return err
	}
	flowId, err := h.config.GetString(config.FlowId)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	h.api.SetApiKey(apiKey)
	remoteEnv, el := h.api.GetRemoteEnvironmentStatus(flowId, remoteEnvId)
	if el != nil {
		return el
	}
	cpapi.PrintPublicEndpoints(h.Stdout, remoteEnv.PublicEndpoints)

	if h.options.dryRun {
		fmt.Fprintln(h.writer, "Dry run mode enabled")
	}

	syncOptions := options.SyncOptions{}
	syncOptions.KubeConfigKey = h.options.environment
	syncOptions.Environment = h.options.environment
	syncOptions.Pod = pod.GetName()
	syncOptions.IndividualFileSyncThreshold = h.options.individualFileSyncThreshold
	syncOptions.RemoteProjectPath = h.options.remoteProjectPath
	syncOptions.DryRun = h.options.dryRun
	syncOptions.Verbose = h.options.rsyncVerbose
	syncOptions.Delete = h.options.delete
	h.syncer.SetOptions(syncOptions)

	dirMonitor.SetLatency(time.Duration(h.options.latency))

	fmt.Fprintf(h.Stdout, "\nDestination Pod: %s\n", pod.GetName())

	observer := sync.GetSyncOnEventObserver(h.syncer)

	return dirMonitor.AnyEventCall(cwd, observer)
}
