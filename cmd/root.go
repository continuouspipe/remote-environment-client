package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "cp-remote",
	Short: "A tool to help with using ContinuousPipe as a remote development environment.",
	Long: `A tool to help with using ContinuousPipe as a remote development environment.

This helps to set up Kubectl, create, build and destroy remote environments and keep files
in sync with the local filesystem.

You will need the following:

 * A ContinuousPipe hosted project with the GitHub integration set up
 * The project checked out locally 
 * The IP address, username and password to use for Kubenetes cluster
 * rsync and fswatch installed locally
 * A https://keen.io write token, project id and event collection name if you want to log usage stats 

Note: if the GitHub repository is not the origin of your checked out project then you will
need to add a remote for that repository. `,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .cp-remote-env-settings.yml in the directory cp-remote is run from.)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	viper.SetConfigName(".cp-remote-env-settings")
	viper.AddConfigPath(pwd)
	if err := viper.ReadInConfig(); err == nil {
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
