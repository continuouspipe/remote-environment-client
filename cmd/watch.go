package cmd

import (
	"bufio"
	"fmt"
	"os"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/spf13/cobra"
)

func NewWatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "watch",
		Aliases: []string{"wa"},
		Short:   "Watch local changes and synchronize with the remote environment",
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
			res, err := handler.AddDefaultCpRemoteExcludeFile(dirWatcher.Exclusions)
			checkErr(err)
			if res == true {
				fmt.Printf("\n%s was missing and has been created with the default ignore settings.\n", sync.SyncExcluded)
			}
			handler.Handle(args, settings, dirWatcher, podsFinder, podsFilter)
		},
	}
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

//if is missing create the SyncExcluded file with default settings
func (handle *WatchHandle) AddDefaultCpRemoteExcludeFile(defaultExclusions []string) (bool, error) {
	//if it exists already skip
	if _, err := os.Stat(sync.SyncExcluded); err == nil {
		return false, nil
	}

	file, err := os.OpenFile(sync.SyncExcluded, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return false, err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range defaultExclusions {
		_, err := w.WriteString(line)
		if err != nil {
			return false, err
		}
		w.WriteString("\n")
	}
	w.Flush()
	return true, nil
}
