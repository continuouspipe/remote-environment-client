package cmd

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/git"
	"github.com/continuouspipe/remote-environment-client/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Setup the remote environment client and settings",
		Long: `This will ask a series of questions to get the details for the project set up.

Your answers will be stored in a .cp-remote-env-settings file in the project root. You
will probably want to add this to your .gitignore file.`,
		Run: func(cmd *cobra.Command, args []string) {
			addApplicationFilesToGitIgnore()

			qp := util.NewQuestionPrompt()
			handler := &SetupHandle{}
			handler.Command = cmd
			handler.qp = qp

			checkErr(handler.Handle(args))

			//fmt.Printf("\nRemote settings written to %s\n", viper.ConfigFileUsed())
			//fmt.Printf("Created the kubernetes config key %s\n", settings.Environment)
			//fmt.Println(kubectlapi.ClusterInfo(settings.Environment))
		},
	}
}

func addApplicationFilesToGitIgnore() {
	gitIgnore, err := git.NewIgnore()
	checkErr(err)
	gitIgnore.AddToIgnore(viper.ConfigFileUsed())
	gitIgnore.AddToIgnore(cplogs.LogDir)
}

type SetupHandle struct {
	Command *cobra.Command
	qp      util.QuestionPrompter
}

func (h *SetupHandle) Handle(args []string) error {

	//yamlWriter := config.NewYamlWriter()

	//request username and password
	username := h.qp.RepeatIfEmpty("What is the continuouspipe username?")
	password := h.qp.RepeatIfEmpty("What is the continuouspipe api-key?")

	//store global configuration settings

	//


	return nil
}

/*func (h *SetupHandle) storeUserSettings(yamlWriter config.Writer) *config.ApplicationSettings {
	team := h.qp.RepeatIfEmpty("What is the continuouspipe team?")
	clusterId := h.qp.RepeatIfEmpty("What is the continuouspipe cluster identifier?")
	projectKey := h.qp.RepeatIfEmpty("What is your Continuous Pipe project key?")
	remoteBranch := h.qp.RepeatIfEmpty("What is the name of the Git branch you are using for your remote environment?")

	settings := &config.ApplicationSettings{
		Team:                team,
		ClusterId:           clusterId,
		ProjectKey:          strings.ToLower(projectKey),
		RemoteBranch:        strings.ToLower(remoteBranch),
		RemoteName:          h.qp.ApplyDefault("What is your github remote name? (defaults to: origin)", "origin"),
		DefaultService:      h.qp.ApplyDefault("What is the default container for the watch, bash, fetch and resync commands? (defaults to: web)", "web"),
		AnybarPort:          h.qp.ReadString("If you want to use AnyBar, please provide a port number e.g 1738 ?"),
		KeenWriteKey:        h.qp.ReadString("What is your keen.io write key? (Optional, only needed if you want to record usage stats)"),
		KeenProjectId:       h.qp.ReadString("What is your keen.io project id? (Optional, only needed if you want to record usage stats)"),
		KeenEventCollection: h.qp.ReadString("What is your keen.io event collection?  (Optional, only needed if you want to record usage stats)"),
		Environment:         strings.ToLower(config.GetEnvironment(projectKey, remoteBranch)),
	}

	tmpl, err := template.New("config").Parse(
		config.Team + ": {{.Team}}\n" +
			config.ClusterId + ": {{.ClusterId}}\n" +
			config.ProjectKey + ": {{.ProjectKey}}\n" +
			config.RemoteBranch + ": {{.RemoteBranch}}\n" +
			config.RemoteName + ": {{.RemoteName}}\n" +
			config.Service + ": {{.DefaultService}}\n" +
			config.AnybarPort + ": {{.AnybarPort}}\n" +
			config.KeenWriteKey + ": {{.KeenWriteKey}}\n" +
			config.KeenProjectId + ": {{.KeenProjectId}}\n" +
			config.KeenEventCollection + ": {{.KeenEventCollection}}\n" +
			config.Environment + ": {{.Environment}}\n" +
			"# Do not change the kubernetes-config-key. If it has been changed please run the setup command again\n" +
			config.KubeConfigKey + ": {{.Environment}}\n")
	checkErr(err)

	yamlWriter.Save(viper.ConfigFileUsed(), tmpl)
	return settings
}

func (h *SetupHandle) IsValidIpAddress(ipAddr string) (bool, error) {
	port := "https"
	cplogs.V(5).Infof("dialling ip address %s:%s", ipAddr, port)
	cplogs.Flush()
	conn, err := net.DialTimeout("tcp", ipAddr+":"+port, 2*time.Second)
	if err != nil {
		cplogs.V(5).Infof("error occurred when dialling ip address %s:%s, error: %s", ipAddr, port, err.Error())
		cplogs.Flush()
		return false, fmt.Errorf("An error occurred when dialling ip address %s:%s, error: %s", ipAddr, port, err.Error())
	}
	conn.Close()
	cplogs.V(5).Infof("connected successfully to ip address %s:%s", ipAddr, port)
	cplogs.Flush()
	return true, nil
}

func applySettingsToCubeCtlConfig(settings *config.ApplicationSettings) {
	kubectlapi.ConfigSetAuthInfo(settings.Environment, settings.Username, settings.Password)
	kubectlapi.ConfigSetCluster(settings.Environment, "kube-proxy.continuouspipe.io:8080", settings.Team, settings.ClusterId)
	kubectlapi.ConfigSetContext(settings.Environment, settings.Username)
}
*/