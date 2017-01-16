package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
)

var fetchCmd = &cobra.Command{
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

		handler := &FetchHandle{cmd}
		podsFinder := pods.NewKubePodsFind()
		podsFilter := pods.NewKubePodsFilter()
		handler.Handle(args, podsFinder, podsFilter)
	},
}

func init() {
	RootCmd.AddCommand(fetchCmd)
}

type FetchHandle struct {
	Command *cobra.Command
}

func (h *FetchHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter) {
	kubeConfigKey := viper.GetString(config.KubeConfigKey)
	environment := viper.GetString(config.Environment)
	service := viper.GetString(config.Service)

	allPods, err := podsFinder.FindAll(kubeConfigKey, environment)
	checkErr(err)

	pod, err := podsFilter.ByService(allPods, service)
	checkErr(err)

	sync.Fetch(kubeConfigKey, environment, pod.GetName())
}
