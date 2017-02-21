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
	settings := config.NewApplicationSettings()
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
			validator := config.NewMandatoryChecker()
			validateConfig(validator, settings)

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
	command.PersistentFlags().StringVarP(&handler.ProjectKey, config.ProjectKey, "p", settings.GetString(config.ProjectKey), "Continuous Pipe project key")
	command.PersistentFlags().StringVarP(&handler.RemoteBranch, config.RemoteBranch, "r", settings.GetString(config.RemoteBranch), "Name of the Git branch you are using for your remote environment")
	command.PersistentFlags().StringVarP(&handler.Service, config.Service, "s", settings.GetString(config.Service), "The service to use (e.g.: web, mysql)")
	command.PersistentFlags().StringVarP(&handler.File, "file", "f", "", "Allows to specify a file that needs to be fetch from the pod")
	command.PersistentFlags().StringVarP(&handler.RemoteProjectPath, "remote-project-path", "a", "/app/", "Specify the absolute path to your project folder, by default set to /app/")
	return command
}

type FetchHandle struct {
	Command           *cobra.Command
	ProjectKey        string
	RemoteBranch      string
	Service           string
	kubeConfigKey     string
	File              string
	RemoteProjectPath string
}

// Complete verifies command line arguments and loads data from the command environment
func (h *FetchHandle) Complete(cmd *cobra.Command, argsIn []string, settingsReader config.Reader) error {
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
	if strings.HasSuffix(h.RemoteProjectPath, "/") == false {
		h.RemoteProjectPath = h.RemoteProjectPath + "/"
	}
	return nil
}

// Validate checks that the provided fetch options are specified.
func (h *FetchHandle) Validate() error {
	if len(strings.Trim(h.ProjectKey, " ")) == 0 {
		return fmt.Errorf("the project key specified is invalid")
	}
	if len(strings.Trim(h.RemoteBranch, " ")) == 0 {
		return fmt.Errorf("the remote branch specified is invalid")
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
	environment := config.GetEnvironment(h.ProjectKey, h.RemoteBranch)

	allPods, err := podsFinder.FindAll(h.kubeConfigKey, environment)
	if err != nil {
		return err
	}

	pod, err := podsFilter.ByService(allPods, h.Service)
	if err != nil {
		return err
	}

	fetcher.SetEnvironment(environment)
	fetcher.SetKubeConfigKey(h.kubeConfigKey)
	fetcher.SetPod(pod.GetName())
	fetcher.SetRemoteProjectPath(h.RemoteProjectPath)
	return fetcher.Fetch(h.File)
}
