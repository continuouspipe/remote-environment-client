package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/errors"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	kubectlcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"os"
	"path/filepath"
)

var localConfigFile string

var usageTemplate = `Usage:{{if .Runnable}}
  {{if .HasAvailableFlags}}{{appendIfNotPresent .UseLine "[flags]"}}{{else}}{{.UseLine}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
  {{ .CommandPath}} [command]{{end}}{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}
{{end}}{{if .HasExample}}

Examples:
{{ .Example }}{{end}}{{ if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .NameAndAliases .NamePadding }}` + "\t" + `{{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{ if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsHelpCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableSubCommands }}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

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
 * rsync and git installed locally

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
	RootCmd.PersistentFlags().StringVar(&localConfigFile, "config", ".cp-remote-settings.yml", "local config file (default is .cp-remote-settings.yml in the directory cp-remote is run from.)")

	RootCmd.AddCommand(NewInitCmd())
	RootCmd.AddCommand(NewBuildCmd())
	RootCmd.AddCommand(NewDestroyCmd())
	RootCmd.AddCommand(NewListPodsCmd())
	RootCmd.AddCommand(NewCheckConnectionCmd())
	RootCmd.AddCommand(NewBashCmd())
	RootCmd.AddCommand(NewExecCmd())
	RootCmd.AddCommand(NewWatchCmd())
	RootCmd.AddCommand(NewFetchCmd())
	RootCmd.AddCommand(NewPushCmd())
	RootCmd.AddCommand(NewSyncCmd())
	RootCmd.AddCommand(NewForwardCmd())
	RootCmd.AddCommand(NewVersionCmd())
	RootCmd.AddCommand(NewCheckUpdatesCmd())
	RootCmd.AddCommand(NewLogsCmd())

	//adding kubectl commands as hidden
	kubeCtlCommand := kubectlcmd.NewKubectlCommand(kubectlcmdutil.NewFactory(nil), os.Stdin, os.Stdout, os.Stderr)
	kubeCtlCommand.Hidden = true

	RootCmd.AddCommand(kubeCtlCommand)

	RootCmd.SetUsageTemplate(usageTemplate)

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	initLocalConfig()
	initGlobalConfig()
	checkLegacyApplicationFile()
	addApplicationFilesToGitIgnore()
}

func initLocalConfig() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	config.C.SetConfigFile(config.LocalConfigType, filepath.Join(pwd, localConfigFile))

	//create the config file if it does not exist
	configFileUsed, err := config.C.ConfigFileUsed(config.LocalConfigType)
	checkErr(err)

	_, err = os.OpenFile(configFileUsed, os.O_RDWR|os.O_CREATE, 0664)
	checkErr(err)
	//load config file
	checkErr(config.C.ReadInConfig(config.LocalConfigType))
}

func initGlobalConfig() {
	homedirPath, err := homedir.Dir()
	checkErr(err)
	globalConfigPath := filepath.Join(homedirPath, ".cp-remote")
	globalConfigFilePath := filepath.Join(globalConfigPath, "config.yml")

	//create the directory
	_ = os.Mkdir(globalConfigPath, 0755)

	//create the global config file
	_, err = os.OpenFile(globalConfigFilePath, os.O_RDWR|os.O_CREATE, 0664)
	checkErr(err)

	//set directory and file path in config
	config.C.SetConfigFile(config.GlobalConfigType, globalConfigFilePath)

	//load config file
	checkErr(config.C.ReadInConfig(config.GlobalConfigType))
}

func checkLegacyApplicationFile() {
	_, err := os.Stat(".cp-remote-env-settings.yml")
	if os.IsNotExist(err) == false {
		color.Red("Warning: you have the '.cp-remote-env-settings.yml' config file which from cp-remote version 0.1.0 it has been replaced by '.cp-remote-settings.yml'.\nPlease remove the old '.cp-remote-env-settings.yml' config file.")
	}
}

func addApplicationFilesToGitIgnore() {
	gitIgnore := config.NewIgnore()
	gitIgnore.File = config.GitIgnore
	logFile, err := config.C.ConfigFileUsed(config.LocalConfigType)
	checkErr(err)
	gitIgnore.AddToIgnore("/" + filepath.Base(logFile))
	gitIgnore.AddToIgnore(cplogs.LogDirName)
}

func validateConfig() {
	valid, missing := config.C.Validate()
	if valid == false {
		errors.ExitWithMessage(fmt.Sprintf("The remote settings file is missing or the require parameters are missing (%v), please run the init command.", missing))

	}
}

func checkErr(err error) {
	errors.CheckErr(err)
}
