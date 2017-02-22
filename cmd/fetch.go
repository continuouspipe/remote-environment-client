package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/benchmark"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/spf13/cobra"
	"strings"
)

var fetchExample = `
# fetch files and folders from the remote pod
cp-remote fe

# fetch files and folders overriding the configuration settings
cp-remote fe -p techup -r dev-user -s web
`

func NewFetchCmd() *cobra.Command {
	settings := config.C
	handler := &FetchHandle{}

	command := &cobra.Command{
		Use:     "fetch",
		Aliases: []string{"fe"},
		Short:   "Sync remote changes to the local filesystem",
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

			benchmark := benchmark.NewCmdBenchmark()
			benchmark.Start("fetch")

			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()
			fetcher := sync.GetFetcher()

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(args, podsFinder, podsFilter, fetcher))

			_, err := benchmark.StopAndLog()
			checkErr(err)
			fmt.Printf("Fetch complete, files and folders retrieved has been logged in %s\n", cplogs.GetLogInfoFile())
			cplogs.Flush()
		},
	}

	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	service, err := settings.GetString(config.Service)
	checkErr(err)

	command.PersistentFlags().StringVarP(&handler.Environment, config.KubeEnvironmentName, "r", environment, "The full remote environment name: project-key-git-branch")
	command.PersistentFlags().StringVarP(&handler.Service, config.Service, "s", service, "The service to use (e.g.: web, mysql)")
	command.PersistentFlags().StringVarP(&handler.File, "file", "f", "", "Allows to specify a file that needs to be fetch from the pod")
	command.PersistentFlags().StringVarP(&handler.RemoteProjectPath, "remote-project-path", "a", "/app/", "Specify the absolute path to your project folder, by default set to /app/")
	return command
}

type FetchHandle struct {
	Command           *cobra.Command
	Environment        string
	Service           string
	File              string
	RemoteProjectPath string
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
	allPods, err := podsFinder.FindAll(h.Environment, h.Environment)
	if err != nil {
		return err
	}

	pod, err := podsFilter.ByService(allPods, h.Service)
	if err != nil {
		return err
	}

	fetcher.SetEnvironment(h.Environment)
	fetcher.SetKubeConfigKey(h.Environment)
	fetcher.SetPod(pod.GetName())
	fetcher.SetRemoteProjectPath(h.RemoteProjectPath)
	return fetcher.Fetch(h.File)
}
