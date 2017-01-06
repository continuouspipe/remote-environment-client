package cmd

import (
	"fmt"

	"bufio"
	envconfig "github.com/alessandrozucca/remote-environment-client/cli/config"
	"github.com/spf13/cobra"
	"os"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Generates the settings file",
	Long: `The setup command uses your answers to generate a settings file .cp-remote-env-settings in the root of the project.
If you need to make changes to the settings you can run the setup command again or you can directly edit the settings.
Note: the kubectl cluster IP address, username and password are not stored in this file. To change these you can run setup again.`,
	Run: func(cmd *cobra.Command, args []string) {
		execute(cmd, args)
	},
}

func execute(cmd *cobra.Command, args []string) {
	//TODO: Put Namespace as function in config package
	//TODO: Repeat questions until the user does not insert a valid string
	//NAMESPACE="$PROJECT_KEY-${REMOTE_BRANCH//\//-}"

	var projectKey string;
	for {
		projectKey := readString("What is your Continuous Pipe project key?")
		if len(projectKey) > 0 {
			break
		}
	}

	remoteBranch := readString("What is the name of the Git branch you are using for your remote environment?")

	config := &envconfig.ConfigData{
		ProjectKey:           projectKey,
		RemoteBranch:         remoteBranch,
		RemoteName:           applyDefault(readString("What is your github remote name? (defaults to: origin)"), "origin"),
		DefaultContainer:     readString("What is the default container for the watch, bash, fetch and resync commands? (Optional)"),
		ClusterIp:            readString("What is the IP of the cluster?"),
		Username:             readString("What is the cluster username?"),
		Password:             readString("What is the cluster password?"),
		AnybarPort:           readString("If you want to use AnyBar, please provide a port number e.g 1738 ?"),
		KeenWriteKey:         readString("What is your keen.io write key? (Optional, only needed if you want to record usage stats)"),
		KeenProjectId:        readString("What is your keen.io project id? (Optional, only needed if you want to record usage stats)"),
		KeenEventCollection:  readString("What is your keen.io event collection?  (Optional, only needed if you want to record usage stats)"),
		Namespace:            projectKey + remoteBranch,
	}

	config.SaveOnDisk()
	fmt.Printf("Remote settings written to %s", envconfig.SettingsFileDir())
}

func init() {
	RootCmd.AddCommand(setupCmd)
}

func readString(q string) string {
	fmt.Print(q, " ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		panic("An error occured when retrieving user input.")
	}
	return input
}

func applyDefault(s string, d string) string {
	if s == "" && d != "" {
		return d
	}
	return s
}