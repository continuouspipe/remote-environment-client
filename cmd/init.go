package cmd

import (
	"encoding/base64"
	"fmt"
	"github.com/continuouspipe/kube-proxy/cplogs"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/git"
	"github.com/spf13/cobra"
	"strings"
	"time"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/util"
)

const InitStateParseSaveToken = "parse-save-token"
const InitStateTriggerBuild = "trigger-build"
const InitStateWaitEnvironmentReady = "wait-environment-ready"
const InitStateApplyEnvironmentSettings = "apply-environment-settings"
const InitStateCompleted = "completed"

const RemoteEnvironmentReadinessProbePeriodSeconds = 60

func NewInitCmd() *cobra.Command {
	settings := config.C
	handler := &InitHandler{}
	handler.config = settings
	handler.qp = util.NewQuestionPrompt()

	command := &cobra.Command{
		Use:     "init [cp-remote-token]",
		Aliases: []string{"in"},
		Short:   "Initialises the remote environment",
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			addApplicationFilesToGitIgnore()

			//Mock base64 token when 5 arguments are passed in
			if len(args) == 5 {
				args = []string{base64.StdEncoding.EncodeToString([]byte(strings.Join(args, ",")))}
			}

			checkErr(handler.Complete(args))
			checkErr(handler.Validate())
			checkErr(handler.Handle())
		},
		Example: portforwardExample,
	}

	remoteName, err := settings.GetString(config.RemoteName)
	checkErr(err)
	command.PersistentFlags().StringVarP(&handler.remoteName, config.KubeEnvironmentName, "r", remoteName, "Override the default remote name (origin)")
	return command
}

func addApplicationFilesToGitIgnore() {
	gitIgnore, err := git.NewIgnore()
	checkErr(err)
	logFile, err := config.C.ConfigFileUsed(config.LocalConfigType)
	checkErr(err)
	gitIgnore.AddToIgnore(logFile)
	gitIgnore.AddToIgnore(cplogs.LogDir)
}

type InitHandler struct {
	command    *cobra.Command
	config     *config.Config
	token      string
	remoteName string
	qp         util.QuestionPrompter
}

// Complete verifies command line arguments and loads data from the command environment
func (i *InitHandler) Complete(argsIn []string) error {
	var err error
	if len(argsIn) > 0 && argsIn[0] != "" {
		i.token = argsIn[0]
		return nil
	}
	if i.remoteName == "" {
		i.remoteName, err = i.config.GetString(config.RemoteName)
		if err != nil {
			return err
		}
		i.config.Set(config.RemoteName, i.remoteName)
	}
	return fmt.Errorf("Invalid token. Please go to continouspipe.io to obtain a valid token.")
}

// Validate checks that the token provided has at least 4 values comma separated
func (i InitHandler) Validate() error {
	if len(strings.Trim(i.remoteName, " ")) == 0 {
		return fmt.Errorf("the remote name specified is invalid")
	}

	decodedToken, err := base64.StdEncoding.DecodeString(i.token)
	if err != nil {
		return fmt.Errorf("Malformed token. Please go to continouspipe.io to obtain a valid token.")
	}
	splitToken := strings.Split(string(decodedToken), ",")
	if len(splitToken) != 5 {
		cplogs.V(5).Infof("Token provided %s has %d parts, expected 4", splitToken, len(splitToken))
		return fmt.Errorf("Malformed token. Please go to continouspipe.io to obtain a valid token.")
	}

	for key, val := range splitToken {
		if val == "" {
			var element string
			switch key {
			case 0:
				element = "api-key"
			case 1:
				element = "remote-environment-id"
			case 2:
				element = "project"
			case 3:
				element = "cp-username"
			case 4:
				element = "git-branch"
			}
			cplogs.V(4).Infof("Element %s is not specified in the token.", element)
			return fmt.Errorf("Malformed token. Please go to continouspipe.io to obtain a valid token.")
		}
	}

	return nil
}

//Handle Executes the initialization
func (i InitHandler) Handle() error {

	currentStatus, err := i.config.GetString(config.InitStatus)
	if err != nil {
		return err
	}

	if currentStatus == InitStateCompleted {
		answer := i.qp.RepeatIfEmpty("The environment is already initialized, do you want to re-initialize? (yes/no)")
		if answer == "no" {
			return nil
		}
		cplogs.V(5).Infoln("The user requested to re-initialize the remote environment")
		//the user want to re-initialize, set the status to empty.
		currentStatus = ""
	}

	var initState initState

	switch currentStatus {
	case "", InitStateParseSaveToken:
		initState = &parseSaveTokenInfo{i.config, i.token}
	case InitStateTriggerBuild:
		initState = NewTriggerBuild(i.config)
	case InitStateWaitEnvironmentReady:
		initState = &waitEnvironmentReady{i.config}
	case InitStateApplyEnvironmentSettings:
		initState = &applyEnvironmentSettings{i.config}
	}

	for initState != nil {
		err := initState.handle()
		if err != nil {
			return err
		}
		initState = initState.next()
	}
	i.config.Set(config.InitStatus, InitStateCompleted)

	return nil
}

type initState interface {
	handle() error
	next() initState
}

type parseSaveTokenInfo struct {
	config *config.Config
	token  string
}

func (p parseSaveTokenInfo) next() initState {
	return NewTriggerBuild(p.config)
}

func (p parseSaveTokenInfo) handle() error {
	p.config.Set(config.InitStatus, InitStateParseSaveToken)

	//we expect the token to have: api-key, remote-environment-id, project, cp-username, git-branch
	splitToken := strings.Split(p.token, ",")
	apiKey := splitToken[0]
	remoteEnvId := splitToken[1]
	project := splitToken[2]
	cpUsername := splitToken[3]
	gitBranch := splitToken[4]

	//check the status of the build on CP to determine if we need to force push or not
	api := cpapi.NewCpApi()
	api.SetApiKey(apiKey)
	cplogs.V(5).Infof("fetching remote environment info for user: %s", cpUsername)
	_, err := api.GetRemoteEnvironment(remoteEnvId)
	if err != nil {
		return err
	}

	cplogs.V(5).Infof("saving parsed token info for user: %s", cpUsername)
	//if there are no errors when fetching the remote environment information we can store the token info
	p.config.Set(config.Username, cpUsername)
	p.config.Set(config.ApiKey, apiKey)
	p.config.Set(config.Project, project)
	p.config.Set(config.RemoteBranch, gitBranch)
	p.config.Set(config.RemoteEnvironmentId, remoteEnvId)
	p.config.Save()
	cplogs.V(5).Infof("saved parsed token info for user: %s", cpUsername)
	cplogs.Flush()
	return nil
}

type triggerBuild struct {
	config   *config.Config
	commit   git.CommitExecutor
	lsRemote git.LsRemoteExecutor
	push     git.PushExecutor
	revList  git.RevListExecutor
	revParse git.RevParseExecutor
}

func NewTriggerBuild(config *config.Config) *triggerBuild {
	return &triggerBuild{
		config,
		git.NewCommit(),
		git.NewLsRemote(),
		git.NewPush(),
		git.NewRevList(),
		git.NewRevParse()}
}

func (p triggerBuild) next() initState {
	return &waitEnvironmentReady{p.config}
}

func (p triggerBuild) handle() error {
	p.config.Set(config.InitStatus, InitStateTriggerBuild)

	apiKey, err := p.config.GetString(config.ApiKey)
	if err != nil {
		return err
	}
	remoteEnvId, err := p.config.GetString(config.RemoteEnvironmentId)
	if err != nil {
		return err
	}
	cpUsername, err := p.config.GetString(config.Username)
	if err != nil {
		return err
	}
	remoteName, err := p.config.GetString(config.RemoteName)
	if err != nil {
		return err
	}
	gitBranch, err := p.config.GetString(config.RemoteBranch)
	if err != nil {
		return err
	}

	api := cpapi.NewCpApi()
	api.SetApiKey(apiKey)
	remoteEnv, err := api.GetRemoteEnvironment(remoteEnvId)
	if err != nil {
		return err
	}

	//if the remote environment is not already building, make sure the remote git branch exists
	//and then trigger a build via api
	if remoteEnv.Status != cpapi.RemoteEnvironmentStatusBuilding {
		cplogs.V(5).Infof("triggering build for the remote environment", cpUsername)

		fmt.Println("Pushing to remote")
		p.createRemoteBranch(remoteName, gitBranch)
		api.RemoteEnvironmentBuild(remoteEnvId)
		fmt.Println("Continuous Pipe will now build your developer environment")
		fmt.Println("You can see when it is complete and find its IP address at https://ui.continuouspipe.io/")
		fmt.Println("Please wait until the build is complete to use any of this tool's other commands.")

	}
	return nil
}

func (i triggerBuild) createRemoteBranch(remoteName string, gitBranch string) error {
	remoteExists, err := i.hasRemote(remoteName, gitBranch)
	if err != nil {
		return err
	}
	if remoteExists == true {
		return nil
	}
	return i.pushLocalBranchToRemote(remoteName, gitBranch)
}

func (i triggerBuild) pushLocalBranchToRemote(remoteName string, gitBranch string) error {
	lbn, err := i.revParse.GetLocalBranchName()
	cplogs.V(5).Infof("local branch name value is %s", lbn)
	if err != nil {
		return err
	}
	i.push.Push(lbn, remoteName, gitBranch)
	return nil
}

func (i triggerBuild) hasRemote(remoteName string, gitBranch string) (bool, error) {
	list, err := i.lsRemote.GetList(remoteName, gitBranch)
	cplogs.V(5).Infof("list of remote branches that matches remote name and branch are %s", list)
	if err != nil {
		return false, err
	}
	if len(list) == 0 {
		return false, err
	}
	return true, err
}

type waitEnvironmentReady struct {
	config *config.Config
}

func (p waitEnvironmentReady) next() initState {
	return &applyEnvironmentSettings{p.config}
}

func (p waitEnvironmentReady) handle() error {
	p.config.Set(config.InitStatus, InitStateWaitEnvironmentReady)

	apiKey, err := p.config.GetString(config.ApiKey)
	if err != nil {
		return err
	}
	remoteEnvId, err := p.config.GetString(config.RemoteEnvironmentId)
	if err != nil {
		return err
	}

	api := cpapi.NewCpApi()
	api.SetApiKey(apiKey)
	remoteEnv, err := api.GetRemoteEnvironment(remoteEnvId)
	if err != nil {
		return err
	}

	//wait until the remote environment has been built
	ticker := time.NewTicker(time.Second * RemoteEnvironmentReadinessProbePeriodSeconds)
	for t := range ticker.C {
		cplogs.V(5).Infoln("environment readiness check at ", t)

		remoteEnv, err = api.GetRemoteEnvironment(remoteEnvId)
		if err != nil {
			return err
		}

		if remoteEnv.Status != cpapi.RemoteEnvironmentStatusNotStarted {
			cplogs.V(5).Infof("re-trying triggering build for the remote environment, status: %s", cpapi.RemoteEnvironmentStatusNotStarted)
			api.RemoteEnvironmentBuild(remoteEnvId)
		}

		if remoteEnv.Status != cpapi.RemoteEnvironmentStatusFailed {
			cplogs.V(5).Infof("remote environment status is %s", cpapi.RemoteEnvironmentStatusFailed)
			return fmt.Errorf("remote environment id %s failed to create.", remoteEnv.RemoteEnvironmentId)
		}

		if remoteEnv.Status != cpapi.RemoteEnvironmentStatusOk {
			cplogs.V(5).Infof("remote environment status is %s", cpapi.RemoteEnvironmentStatusOk)
			continue
		}
		cplogs.Flush()
	}
	return nil
}

type applyEnvironmentSettings struct {
	config *config.Config
}

func (p applyEnvironmentSettings) next() initState {
	return nil
}

func (p applyEnvironmentSettings) handle() error {
	p.config.Set(config.InitStatus, InitStateApplyEnvironmentSettings)

	apiKey, err := p.config.GetString(config.ApiKey)
	if err != nil {
		return err
	}
	remoteEnvId, err := p.config.GetString(config.RemoteEnvironmentId)
	if err != nil {
		return err
	}

	api := cpapi.NewCpApi()
	api.SetApiKey(apiKey)

	remoteEnv, err := api.GetRemoteEnvironment(remoteEnvId)
	if err != nil {
		return err
	}

	cplogs.V(5).Infof("saving remote environment info for environment name: %s, environment id: %s", remoteEnv.KubeEnvironmentName, remoteEnv.RemoteEnvironmentId)
	//the environment has been built, so save locally the settings received from the server
	p.config.Set(config.RemoteEnvironmentConfigModifiedAt, remoteEnv.ModifiedAt)
	p.config.Set(config.ClusterIdentifier, remoteEnv.ClusterIdentifier)
	p.config.Set(config.AnybarPort, remoteEnv.AnyBarPort)
	p.config.Set(config.KubeEnvironmentName, remoteEnv.KubeEnvironmentName)
	p.config.Set(config.KeenEventCollection, remoteEnv.KeenEventCollection)
	p.config.Set(config.KeenProjectId, remoteEnv.KeenId)
	p.config.Set(config.KeenWriteKey, remoteEnv.KeenWriteKey)
	p.config.Save()
	cplogs.V(5).Infoln("saved remote environment info")
	cplogs.Flush()

	err = p.applySettingsToCubeCtlConfig()
	if err != nil {
		return err
	}
	p.config.Save()
	cplogs.Flush()
	return nil
}

func (i applyEnvironmentSettings) applySettingsToCubeCtlConfig() error {
	environment, err := i.config.GetString(config.KubeEnvironmentName)
	if err != nil {
		return err
	}
	username, err := i.config.GetString(config.Username)
	if err != nil {
		return err
	}
	apiKey, err := i.config.GetString(config.ApiKey)
	if err != nil {
		return err
	}
	project, err := i.config.GetString(config.Project)
	if err != nil {
		return err
	}
	clusterId, err := i.config.GetString(config.ClusterIdentifier)
	if err != nil {
		return err
	}

	kubectlapi.ConfigSetAuthInfo(environment, username, apiKey)
	kubectlapi.ConfigSetCluster(environment, "kube-proxy-staging.continuouspipe.io:8080", project, clusterId)
	kubectlapi.ConfigSetContext(environment, username)

	localConfigFile, err := i.config.ConfigFileUsed(config.LocalConfigType)
	if err != nil {
		return err
	}
	globalConfigFile, err := i.config.ConfigFileUsed(config.GlobalConfigType)
	if err != nil {
		return err
	}

	fmt.Printf("\nRemote settings written to %s\n", localConfigFile)
	fmt.Printf("\nRemote settings written to %s\n", globalConfigFile)
	fmt.Printf("Created the kubernetes config key %s\n", environment)
	fmt.Println(kubectlapi.ClusterInfo(environment))
	return nil
}
