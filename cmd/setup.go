package cmd

import (
	"fmt"
	"bufio"
	"os"
	"strings"

	envconfig "github.com/continuouspipe/remote-environment-client/config"
	"github.com/spf13/cobra"
)

type reader func(string) string

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the remote environment client and settings",
	Long: `This will ask a series of questions to get the details for the project set up. 
	    
Your answers will be stored in a .cp-remote-env-settings file in the project root. You 
will probably want to add this to your .gitignore file.`,
	Run: func(cmd *cobra.Command, args []string) {
		execute(cmd, args)
	},
}

func execute(cmd *cobra.Command, args []string) {
	projectKey := repeatIfEmpty(readString, "What is your Continuous Pipe project key?")
	remoteBranch := repeatIfEmpty(readString, "What is the name of the Git branch you are using for your remote environment?")

	namespace := strings.Replace(remoteBranch, "/", "-", -1)
	namespace = strings.Replace(namespace, "\\", "-", -1)
	namespace = projectKey + "-" + namespace

	config := &envconfig.ConfigData{
		ProjectKey:           projectKey,
		RemoteBranch:         remoteBranch,
		RemoteName:           applyDefault(readString, "What is your github remote name? (defaults to: origin)", "origin"),
		DefaultContainer:     readString("What is the default container for the watch, bash, fetch and resync commands? (Optional)"),
		ClusterIp:            repeatIfEmpty(readString, "What is the IP of the cluster?"),
		Username:             repeatIfEmpty(readString, "What is the cluster username?"),
		Password:             repeatIfEmpty(readString, "What is the cluster password?"),
		AnybarPort:           readString("If you want to use AnyBar, please provide a port number e.g 1738 ?"),
		KeenWriteKey:         readString("What is your keen.io write key? (Optional, only needed if you want to record usage stats)"),
		KeenProjectId:        readString("What is your keen.io project id? (Optional, only needed if you want to record usage stats)"),
		KeenEventCollection:  readString("What is your keen.io event collection?  (Optional, only needed if you want to record usage stats)"),
		Namespace:            projectKey + remoteBranch,
	}

	config.SaveOnDisk()
	fmt.Printf("\nRemote settings written to %s\n", envconfig.SettingsFileDir())
}

func init() {
	RootCmd.AddCommand(setupCmd)
}

func readString(q string) string {
	fmt.Print(q, " ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func applyDefault(ask reader, question string, predef string) string {
	res := ask(question)

	if res == "" && predef != "" {
		return predef
	}
	return res
}

func repeatIfEmpty(r reader, question string) string {
	var res string
	for {
		res = r(question)
		if len(res) > 0 {
			break
		}
	}
	return res
}
