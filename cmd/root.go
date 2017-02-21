package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/mitchellh/go-homedir"
	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	kubectlcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"os"
	"runtime/debug"
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
	RootCmd.PersistentFlags().StringVar(&localConfigFile, "config", ".cp-remote-settings.yml", "local config file (default is .cp-remote-settings.yml in the directory cp-remote is run from.)")

	RootCmd.AddCommand(NewBashCmd())
	RootCmd.AddCommand(NewBuildCmd())
	RootCmd.AddCommand(NewCheckConnectionCmd())
	RootCmd.AddCommand(NewCheckUpdatesCmd())
	RootCmd.AddCommand(NewDestroyCmd())
	RootCmd.AddCommand(NewExecCmd())
	RootCmd.AddCommand(NewFetchCmd())
	RootCmd.AddCommand(NewForwardCmd())
	RootCmd.AddCommand(NewInitCmd())
	RootCmd.AddCommand(NewVersionCmd())
	RootCmd.AddCommand(NewWatchCmd())

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
}

func initLocalConfig() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	config.C.SetConfigFile(config.LocalConfigType, pwd + string(os.PathSeparator) + localConfigFile)

	//create the config file if it does not exist
	configFileUsed, err := config.C.ConfigFileUsed(config.LocalConfigType)
	checkErr(err)

	_, err = os.OpenFile(configFileUsed, os.O_RDWR|os.O_CREATE, 0664)
	checkErr(err)
	//load config file
	checkErr(config.C.ReadInConfig(config.LocalConfigType))
}

func initGlobalConfig() {
	homedir, err := homedir.Dir()
	checkErr(err)
	globalConfigPath := homedir + string(os.PathSeparator) + ".cp-remote" + string(os.PathSeparator)
	globalConfigName := "config.yml"

	//create the directory
	_ = os.Mkdir(globalConfigPath, 0755)

	//create the global config file
	_, err = os.OpenFile(globalConfigPath+globalConfigName, os.O_RDWR|os.O_CREATE, 0664)
	checkErr(err)

	//set directory and file path in config
	config.C.SetConfigFile(config.GlobalConfigType, globalConfigPath + globalConfigName)

	//load config file
	checkErr(config.C.ReadInConfig(config.GlobalConfigType))
}

func validateConfig() {
	valid, missing := config.C.Validate()
	if valid == false {
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

	stack := debug.Stack()
	cplogs.V(4).Info(string(stack[:]))
	color.Unset()
	cplogs.Flush()
	os.Exit(1)
}
