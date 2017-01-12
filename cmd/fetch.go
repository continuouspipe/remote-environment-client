package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/sync"
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
		handler := &FetchHandle{cmd}
		handler.Handle(args)
	},
}

func init() {
	RootCmd.AddCommand(fetchCmd)
}

type FetchHandle struct {
	Command *cobra.Command
}

func (h *FetchHandle) Handle(args []string) {
	validateConfig()

	kubeConfigKey := viper.GetString(config.KubeConfigKey)
	environment := viper.GetString(config.Environment)
	service := viper.GetString(config.Service)

	pod, err := kubectlapi.FindPodByService(kubeConfigKey, environment, service)
	checkErr(err)

	sync.Fetch(kubeConfigKey, environment, service, pod.GetName())
}
