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
)

const InitStatusStarted = "started"
const InitStatusCompleted = "completed"

const RemoteEnvironmentReadinessProbePeriodSeconds = 60

func NewInitCmd() *cobra.Command {
	settings := config.C
	handler := &InitHandler{}
	handler.config = settings
	handler.commit = git.NewCommit()
	handler.lsRemote = git.NewLsRemote()
	handler.push = git.NewPush()
	handler.revList = git.NewRevList()
	handler.revParse = git.NewRevParse()

	command := &cobra.Command{
		Use:     "init [cp-remote-token]",
		Aliases: []string{"in"},
		Short:   "Initialises the remote environment",
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {

			//Mock base64 token when 4 arguments are passed in
			if len(args) == 4 {
				args = []string{base64.StdEncoding.EncodeToString([]byte(strings.Join(args, ",")))}
			}

			checkErr(handler.Complete(args))
			checkErr(handler.Validate())
			checkErr(handler.Handle())
		},
		Example: portforwardExample,
	}
	return command
}

type InitHandler struct {
	command  *cobra.Command
	config   *config.Config
	commit   git.CommitExecutor
	lsRemote git.LsRemoteExecutor
	push     git.PushExecutor
	revList  git.RevListExecutor
	revParse git.RevParseExecutor
	token    string
}

// Complete verifies command line arguments and loads data from the command environment
func (i *InitHandler) Complete(argsIn []string) error {

	fmt.Println(len(argsIn))
	if len(argsIn) > 0 && argsIn[0] != "" {
		fmt.Println(argsIn[0])
		i.token = argsIn[0]
		return nil
	}
	return fmt.Errorf("Invalid token. Please go to continouspipe.io to obtain a valid token.")
}

// Validate checks that the token provided has at least 4 values comma separated
func (i InitHandler) Validate() error {
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

	initStatus, err := i.config.GetString(config.InitStatus)
	if err != nil {
		return err
	}

	//if the init status is not set, is the first time init has been called on the project.
	if initStatus == "" {
		return i.initializeNewRemoteEnvironment()
	}

	//init status started
	//check what's the situation: do we have api-key? is it building on CP? do we have git-branch to push if necessary?

	//init status completed
	//do we ask the user if he want to rebuild using this new init token?
	return nil
}

func (i InitHandler) initializeNewRemoteEnvironment() error {
	//set the init-status as "STARTED"
	i.config.Set(config.InitStatus, InitStatusStarted)

	//we expect the token to have: api-key, remote-environment-id, cp-username, git-branch
	splitToken := strings.Split(i.token, ",")
	apiKey := splitToken[0]
	remoteEnvironmentId := splitToken[1]
	project := splitToken[2]
	cpUsername := splitToken[3]
	gitBranch := splitToken[4]

	//check the status of the build on CP to determine if we need to force push or not
	api := cpapi.NewCpApi()
	api.SetApiKey(apiKey)
	cplogs.V(5).Infof("fetching remote environment info for user: %s", cpUsername)
	remoteEnvironment, err := api.GetRemoteEnvironment(remoteEnvironmentId)
	if err != nil {
		return err
	}

	//if there are no errors when fetching the remote environment information we can store the token info
	i.config.Set(config.Username, cpUsername)
	i.config.Set(config.ApiKey, apiKey)
	i.config.Set(config.Project, project)
	i.config.Set(config.RemoteBranch, gitBranch)
	i.config.Set(config.RemoteEnvironmentId, remoteEnvironmentId)
	i.config.Save()

	//if the remote environment is not already building, make sure the remote git branch exists
	//and then trigger a build via api
	if remoteEnvironment.Status != cpapi.RemoteEnvironmentStatusBuilding {
		remoteName, err := i.config.GetString(config.RemoteName)
		if err != nil {
			return err
		}
		i.createRemoteBranch(remoteName, gitBranch)
		api.RemoteEnvironmentBuild(remoteEnvironmentId)
	}

	//wait until the remote environment has been built
	ticker := time.NewTicker(time.Second * RemoteEnvironmentReadinessProbePeriodSeconds)
	for t := range ticker.C {
		cplogs.V(5).Infoln("environment readiness check at ", t)

		remoteEnvironment, err = api.GetRemoteEnvironment(remoteEnvironmentId)
		if err != nil {
			return err
		}

		if remoteEnvironment.Status != cpapi.RemoteEnvironmentStatusNotStarted {
			api.RemoteEnvironmentBuild(remoteEnvironmentId)
		}

		if remoteEnvironment.Status != cpapi.RemoteEnvironmentStatusFailed {
			return fmt.Errorf("remote environment id %s failed to create.", remoteEnvironment.RemoteEnvironmentId)
		}

		if remoteEnvironment.Status != cpapi.RemoteEnvironmentStatusOk {
			continue
		}
	}

	//the environment has been built, so save locally the settings received from the server
	i.config.Set(config.RemoteEnvironmentConfigModifiedAt, remoteEnvironment.ModifiedAt)
	i.config.Set(config.ClusterIdentifier, remoteEnvironment.ClusterIdentifier)
	i.config.Set(config.KubeEnvironmentName, remoteEnvironment.KubeEnvironmentName)
	i.config.Set(config.KeenEventCollection, remoteEnvironment.KeenEventCollection)
	i.config.Set(config.KeenProjectId, remoteEnvironment.KeenId)
	i.config.Set(config.KeenWriteKey, remoteEnvironment.KeenWriteKey)
	i.config.Save()

	return i.applySettingsToCubeCtlConfig()
}

func (i InitHandler) applySettingsToCubeCtlConfig() error {
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
}

func (i InitHandler) createRemoteBranch(remoteName string, gitBranch string) error {
	remoteExists, err := i.hasRemote(remoteName, gitBranch)
	if err != nil {
		return err
	}
	if remoteExists == true {
		return nil
	}
	return i.pushLocalBranchToRemote(remoteName, gitBranch)
}

func (i InitHandler) pushLocalBranchToRemote(remoteName string, gitBranch string) error {
	lbn, err := i.revParse.GetLocalBranchName()
	cplogs.V(5).Infof("local branch name value is %s", lbn)
	if err != nil {
		return err
	}
	i.push.Push(lbn, remoteName, gitBranch)
	return nil
}

func (i InitHandler) hasRemote(remoteName string, gitBranch string) (bool, error) {
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
