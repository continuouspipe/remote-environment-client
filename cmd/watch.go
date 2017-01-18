package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch local changes and synchronize with the remote environment",
	Long: `The watch command will sync changes you make locally to a container that's part
of the remote environment. This will use the default container specified during
setup but you can specify another container to sync with.`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("Watching for changes. Quit anytime with Ctrl-C.")

		settings := config.NewApplicationSettings()
		validator := config.NewMandatoryChecker()
		validateConfig(validator, settings)
		dirWatcher := sync.GetRecursiveDirectoryMonitor()
		podsFinder := pods.NewKubePodsFind()
		podsFilter := pods.NewKubePodsFilter()

		handler := &WatchHandle{cmd}
		handler.Handle(args, settings, dirWatcher, podsFinder, podsFilter)
	},
}

func init() {
	RootCmd.AddCommand(watchCmd)
}

type WatchHandle struct {
	Command *cobra.Command
}

func (h *WatchHandle) Handle(args []string, settings config.Reader, recursiveDirWatcher sync.DirectoryMonitor, podsFinder pods.Finder, podsFilter pods.Filter) {
	kubeConfigKey := settings.GetString(config.KubeConfigKey)
	environment := settings.GetString(config.Environment)
	service := settings.GetString(config.Service)

	allPods, err := podsFinder.FindAll(kubeConfigKey, environment)
	checkErr(err)

	pod, err := podsFilter.ByService(allPods, service)
	checkErr(err)

	observer := sync.GetDirectoryEventSyncAll()
	observer.KubeConfigKey = kubeConfigKey
	observer.Environment = environment
	observer.Pod = *pod
	err = recursiveDirWatcher.AnyEventCall(".", observer)
	checkErr(err)
}
