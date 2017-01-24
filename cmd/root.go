package cmd

import (
	"fmt"
	"os"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	kubectlcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

var cfgFile string

type Handler interface {
	Handle(cmd *cobra.Command, args []string)
}

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
need to add a remote for that repository.`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".cp-remote-env-settings.yml", "config file (default is .cp-remote-env-settings.yml in the directory cp-remote is run from.)")
	cobra.OnInitialize(initConfig)

	RootCmd.AddCommand(NewBashCmd())
	RootCmd.AddCommand(NewBuildCmd())
	RootCmd.AddCommand(NewCheckConnectionCmd())
	RootCmd.AddCommand(NewCheckUpdatesCmd())
	RootCmd.AddCommand(NewDestroyCmd())
	RootCmd.AddCommand(NewExecCmd())
	RootCmd.AddCommand(NewFetchCmd())
	RootCmd.AddCommand(NewForwardCmd())
	RootCmd.AddCommand(NewSetupCmd())
	RootCmd.AddCommand(NewVersionCmd())
	RootCmd.AddCommand(NewWatchCmd())

	//adding kubectl commands as hidden
	kubeCtlCommand := kubectlcmd.NewKubectlCommand(kubectlcmdutil.NewFactory(nil), os.Stdin, os.Stdout, os.Stderr)
	kubeCtlCommand.Hidden = true
	RootCmd.AddCommand(kubeCtlCommand)
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

	viper.AddConfigPath(pwd)
	checkErr(viper.ReadInConfig())
}

func validateConfig(validator config.Validator, reader config.Reader) {
	i, missing := validator.Validate(reader)
	if i > 0 {
		exitWithMessage(fmt.Sprintf("The remote settings file is missing or the require parameters are missing (%v), please run the setup command.", missing))
	}
}

func checkErr(err error) {
	if err != nil {
		exitWithMessage(err.Error())
	}
}

func exitWithMessage(message string) {
	color.Set(color.FgRed)
	fmt.Println("ERROR: " + message)
	color.Unset()
	cplogs.Flush()
	os.Exit(1)
}
