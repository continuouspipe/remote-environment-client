package cmd

import (
	"github.com/spf13/cobra"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/sync"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch local changes and synchronize with the remote environment",
	Long: `The watch command will sync changes you make locally to a container that's part
of the remote environment. This will use the default container specified during
setup but you can specify another container to sync with.`,
	Run: func(cmd *cobra.Command, args []string) {
		settings := config.NewApplicationSettings()
		validator := config.NewMandatoryChecker()
		validateConfig(validator, settings)
		dirWatcher := sync.GetDirectoryWatch()

		handler := &WatchHandle{cmd}
		handler.Handle(args, settings, dirWatcher)
	},
}

func init() {
	RootCmd.AddCommand(watchCmd)
	//watchCmd.PersistentFlags().StringP("container", "c", "", "The container to use")
	//watchCmd.PersistentFlags().StringP("environment", "e", "", "The environment to use")
}

type WatchHandle struct {
	Command *cobra.Command
}

func (h *WatchHandle) Handle(args []string, settings config.Reader, dirWatcher sync.DirectoryWatchEventCaller) {
	//viper.BindPFlag("container", h.Command.PersistentFlags().Lookup("container"))
	//viper.BindPFlag("environment", h.Command.PersistentFlags().Lookup("environment"))
	//fmt.Println("The container is set to: " + viper.GetString("container"))
	//fmt.Println("The environment is set to: " + viper.GetString("environment"))

	observer := sync.GetDirectoryEventSyncAll()
	dirWatcher.AnyEventCall(".", observer)
}
