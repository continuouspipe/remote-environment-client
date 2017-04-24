package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/benchmark"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/continuouspipe/remote-environment-client/sync/options"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

var fetchExample = fmt.Sprintf(`
# fetch files and folders from the remote pod
%[1]s fe
# fetch files and folders to the remote pod specifying the environment
%[1]s fe -e techup-dev-user -s web
`, config.AppName)

func NewFetchCmd() *cobra.Command {
	settings := config.C
	handler := &FetchHandle{}
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	handler.writer = os.Stdout

	command := &cobra.Command{
		Use:     "fetch",
		Aliases: []string{"fe"},
		Short:   "Fetches remote changes to the local filesystem",
		Example: fetchExample,
		Long: `When the remote environment is rebuilt it may contain changes that you do not
have on the local filesystem. For example, for a PHP project part of building the remote
environment could be installing the vendors using composer. Any new or updated vendors would
be on the remote environment but not on the local filesystem which would cause issues, such as
autocomplete in your IDE not working correctly.

The fetch command will copy changes from the remote to the local filesystem. This will resync
with the default container specified during setup but you can specify another container.`,
		Run: func(cmd *cobra.Command, args []string) {
			validateConfig()

			fmt.Println("Fetch in progress")

			b := benchmark.NewCmdBenchmark()
			b.Start("fetch")

			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()
			fetcher := sync.GetFetcher()

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(args, podsFinder, podsFilter, fetcher))

			_, err := b.StopAndLog()
			checkErr(err)
			fmt.Printf("Fetch complete, files and folders retrieved has been logged in %s\n", cplogs.GetLogInfoFile())
			cplogs.Flush()
		},
	}

	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	service, err := settings.GetString(config.Service)
	checkErr(err)

	command.PersistentFlags().StringVarP(&handler.Environment, config.KubeEnvironmentName, "e", environment, "The full remote environment name")
	command.PersistentFlags().StringVarP(&handler.Service, config.Service, "s", service, "The service to use (e.g.: web, mysql)")
	command.PersistentFlags().StringVarP(&handler.File, "file", "f", "", "Allows to specify a file that needs to be fetch from the pod")
	command.PersistentFlags().StringVarP(&handler.RemoteProjectPath, "remote-project-path", "a", "/app/", "Specify the absolute path to your project folder, by default set to /app/")
	command.PersistentFlags().BoolVar(&handler.rsyncVerbose, "rsync-verbose", false, "Allows to use rsync in verbose mode and debug issues with exclusions")
	command.PersistentFlags().BoolVar(&handler.dryRun, "dry-run", false, "Show what would have been transferred")
	return command
}

type FetchHandle struct {
	Command           *cobra.Command
	Environment       string
	Service           string
	File              string
	RemoteProjectPath string
	kubeCtlInit       kubectlapi.KubeCtlInitializer
	rsyncVerbose      bool
	dryRun            bool
	writer            io.Writer
}

// Complete verifies command line arguments and loads data from the command environment
func (h *FetchHandle) Complete(cmd *cobra.Command, argsIn []string, settings *config.Config) error {
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

// Validate checks that the provided fetch options are specified.
func (h *FetchHandle) Validate() error {
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

// Copies all the files and folders from the remote development environment into the current directory
func (h *FetchHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter, fetcher sync.Fetcher) error {
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return nil
	}

	allPods, err := podsFinder.FindAll(user, apiKey, addr, h.Environment)
	if err != nil {
		return err
	}

	pod := podsFilter.List(*allPods).ByService(h.Service).ByStatus("Running").First()
	if pod == nil {
		return fmt.Errorf(fmt.Sprintf(msgs.NoActivePodsFoundForSpecifiedServiceName, h.Service))
	}

	if h.dryRun {
		fmt.Fprintln(h.writer, "Dry run mode enabled")
	}

	syncOptions := options.SyncOptions{}
	syncOptions.Verbose = h.rsyncVerbose
	syncOptions.Environment = h.Environment
	syncOptions.KubeConfigKey = h.Environment
	syncOptions.Pod = pod.GetName()
	syncOptions.RemoteProjectPath = h.RemoteProjectPath
	syncOptions.DryRun = h.dryRun
	fetcher.SetOptions(syncOptions)
	return fetcher.Fetch(h.File)
}
