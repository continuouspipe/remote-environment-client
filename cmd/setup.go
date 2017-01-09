package cmd

import (
	"fmt"
	"strings"

	envconfig "github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the remote environment client and settings",
	Long: `This will ask a series of questions to get the details for the project set up. 
	    
Your answers will be stored in a .cp-remote-env-settings file in the project root. You 
will probably want to add this to your .gitignore file.`,
	Run: func(cmd *cobra.Command, args []string) {
		qp := util.NewQuestionPrompt()
		yamlWriter := envconfig.NewYamlWriter()
		runSetupCmd(qp, yamlWriter)
		fmt.Printf("\nRemote settings written to %s\n", viper.ConfigFileUsed())
	},
}

func runSetupCmd(qp util.QuestionPrompter, yamlWriter envconfig.Writer) {
	projectKey := qp.RepeatIfEmpty("What is your Continuous Pipe project key?")
	remoteBranch := qp.RepeatIfEmpty("What is the name of the Git branch you are using for your remote environment?")

	namespace := strings.Replace(remoteBranch, "/", "-", -1)
	namespace = strings.Replace(namespace, "\\", "-", -1)
	namespace = projectKey + "-" + namespace

	settings := &envconfig.ApplicationSettings{
		ProjectKey:          projectKey,
		RemoteBranch:        remoteBranch,
		RemoteName:          qp.ApplyDefault("What is your github remote name? (defaults to: origin)", "origin"),
		DefaultContainer:    qp.ReadString("What is the default container for the watch, bash, fetch and resync commands? (Optional)"),
		ClusterIp:           qp.RepeatIfEmpty("What is the IP of the cluster?"),
		Username:            qp.RepeatIfEmpty("What is the cluster username?"),
		Password:            qp.RepeatIfEmpty("What is the cluster password?"),
		AnybarPort:          qp.ReadString("If you want to use AnyBar, please provide a port number e.g 1738 ?"),
		KeenWriteKey:        qp.ReadString("What is your keen.io write key? (Optional, only needed if you want to record usage stats)"),
		KeenProjectId:       qp.ReadString("What is your keen.io project id? (Optional, only needed if you want to record usage stats)"),
		KeenEventCollection: qp.ReadString("What is your keen.io event collection?  (Optional, only needed if you want to record usage stats)"),
		Namespace:           namespace,
	}

	yamlWriter.Save(settings)
}

func init() {
	RootCmd.AddCommand(setupCmd)
}
