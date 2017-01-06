package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch local changes and synchronize with the remote environment",
	Long: `The watch command will sync changes you make locally to a container that's part
of the remote environment. This will use the default container specified during
setup but you can specify another container to sync with.`,
	Run: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("container", cmd.PersistentFlags().Lookup("container"))
		viper.BindPFlag("environment", cmd.PersistentFlags().Lookup("environment"))

		fmt.Println("watch called")
		fmt.Println("The container is set to: " + viper.GetString("container"))
		fmt.Println("The environment is set to: " + viper.GetString("environment"))
	},
}

func init() {
	RootCmd.AddCommand(watchCmd)

	watchCmd.PersistentFlags().StringP("container", "c", "", "The container to use")
	watchCmd.PersistentFlags().StringP("environment", "e", "", "The environment to use")
}
