package cmd

import (
	"github.com/continuouspipe/remote-environment-client/benchmark"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/spf13/cobra"
	"fmt"
	"github.com/continuouspipe/remote-environment-client/cplogs"
)

func NewFetchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fetch",
		Short: "Sync remote changes to the local filesystem",
		Long: `When the remote environment is rebuilt it may contain changes that you do not
have on the local filesystem. For example, for a PHP project part of building the remote
environment could be installing the vendors using composer. Any new or updated vendors would
be on the remote environment but not on the local filesystem which would cause issues, such as
autocomplete in your IDE not working correctly.

The fetch command will copy changes from the remote to the local filesystem. This will resync
with the default container specified during setup but you can specify another container.`,
		Run: func(cmd *cobra.Command, args []string) {
			settings := config.NewApplicationSettings()
			validator := config.NewMandatoryChecker()
			validateConfig(validator, settings)

			fmt.Println("Fetch in progress")

			benchmark := benchmark.NewCmdBenchmark()
			benchmark.Start("fetch")

			handler := &FetchHandle{cmd}
			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()
			rsyncFetch := sync.NewRsyncFetch()

			err := handler.Handle(args, settings, podsFinder, podsFilter, rsyncFetch)
			checkErr(err)
			_, err = benchmark.StopAndLog()
			checkErr(err)
			fmt.Printf("Fetch complete, files and folders retrieved has been logged in %s\n", cplogs.GetLogInfoFile())
			cplogs.Flush()
		},
	}
}

type FetchHandle struct {
	Command *cobra.Command
}

func (h *FetchHandle) Handle(args []string, settings config.Reader, podsFinder pods.Finder, podsFilter pods.Filter, fetcher sync.Fetcher) error {
	kubeConfigKey := settings.GetString(config.KubeConfigKey)
	environment := settings.GetString(config.Environment)
	service := settings.GetString(config.Service)

	allPods, err := podsFinder.FindAll(kubeConfigKey, environment)
	checkErr(err)

	pod, err := podsFilter.ByService(allPods, service)
	checkErr(err)

	return fetcher.Fetch(kubeConfigKey, environment, pod.GetName())
}
