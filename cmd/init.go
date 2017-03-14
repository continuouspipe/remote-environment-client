package cmd

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/git"
	"github.com/continuouspipe/remote-environment-client/initialization"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/services"
	"github.com/continuouspipe/remote-environment-client/util"
	"github.com/spf13/cobra"
	"io"
	"os"
)

const initStateParseSaveToken = "parse-save-token"
const initStateTriggerBuild = "trigger-build"
const initStateWaitEnvironmentReady = "wait-environment-ready"
const initStateApplyEnvironmentSettings = "apply-environment-settings"
const initStateApplyDefaultService = "apply-default-service"
const initStateCompleted = "completed"

const remoteEnvironmentReadinessProbePeriodSeconds = 30

//NewInitCmd Initialises the remote environment
func NewInitCmd() *cobra.Command {
	settings := config.C
	handler := &initHandler{}
	handler.api = cpapi.NewCpApi()
	handler.config = settings
	handler.qp = util.NewQuestionPrompt()

	command := &cobra.Command{
		Use:     "init",
		Aliases: []string{"in", "setup"},
		Short:   "Initialises the remote environment",
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			//Mock base64 token when 5 arguments are passed in
			if len(args) == 5 {
				args = []string{base64.StdEncoding.EncodeToString([]byte(strings.Join(args, ",")))}
			}

			checkErr(handler.Complete(args))
			checkErr(handler.Validate())
			checkErr(handler.Handle())
		},
	}

	remoteName, err := settings.GetString(config.RemoteName)
	checkErr(err)
	command.PersistentFlags().StringVar(&handler.remoteName, config.RemoteName, remoteName, "Override the default remote name (origin)")
	command.PersistentFlags().BoolVarP(&handler.reset, "reset", "r", false, "With reset flag set to true, init will not attempt to restore interrupted initializations")
	return command
}

type initHandler struct {
	command    *cobra.Command
	config     config.ConfigProvider
	token      string
	remoteName string
	reset      bool
	qp         util.QuestionPrompter
	api        cpapi.CpApiProvider
}

// Complete verifies command line arguments and loads data from the command environment
func (i *initHandler) Complete(argsIn []string) error {
	var err error
	if len(argsIn) > 0 && argsIn[0] != "" {
		decodedToken, err := base64.StdEncoding.DecodeString(argsIn[0])
		if err != nil {
			return fmt.Errorf("Malformed token. Please go to continouspipe.io to obtain a valid token")
		}
		i.token = string(decodedToken)
		return nil
	}
	if i.remoteName == "" {
		i.remoteName, err = i.config.GetString(config.RemoteName)
		if err != nil {
			return err
		}
		i.config.Set(config.RemoteName, i.remoteName)
	}
	return fmt.Errorf("Invalid token. Please go to continouspipe.io to obtain a valid token")
}

// Validate checks that the token provided has at least 5 values comma separated
func (i initHandler) Validate() error {
	splitToken := strings.Split(string(i.token), ",")
	if len(splitToken) != 5 {
		cplogs.V(5).Infof("Token provided %s has %d parts, expected 4", splitToken, len(splitToken))
		return fmt.Errorf("Malformed token. Please go to continouspipe.io to obtain a valid token")
	}

	for key, val := range splitToken {
		if val == "" {
			var element string
			switch key {
			case 0:
				element = "project"
			case 1:
				element = "remote-environment-id"
			case 2:
				element = "api-key"
			case 3:
				element = "cp-username"
			case 4:
				element = "git-branch"
			}
			cplogs.V(4).Infof("element %s is not specified in the token.", element)
			return fmt.Errorf("Malformed token. Please go to continouspipe.io to obtain a valid token")
		}
	}

	return nil
}

//Handle Executes the initialization
func (i initHandler) Handle() error {
	currentStatus, err := i.config.GetString(config.InitStatus)
	if err != nil {
		return err
	}

	if currentStatus == initStateCompleted {
		answer := i.qp.RepeatIfEmpty("The environment is already initialized, do you want to re-initialize? (yes/no)")
		if answer == "no" {
			return nil
		}
		cplogs.V(5).Infoln("The user requested to re-initialize the remote environment")
		//the user want to re-initialize, set the status to empty.
		currentStatus = ""
	}

	if i.reset == true {
		//the user want to re-initialize, set the status to empty.
		currentStatus = ""
	}

	var initState initialization.InitState

	switch currentStatus {
	case "", initStateParseSaveToken:
		initState = &parseSaveTokenInfo{i.config, i.token, cpapi.NewCpApi()}
	case initStateTriggerBuild:
		initState = newTriggerBuild()
	case initStateWaitEnvironmentReady:
		initState = newWaitEnvironmentReady()
	case initStateApplyEnvironmentSettings:
		initState = newApplyEnvironmentSettings()
	case initStateApplyDefaultService:
		initState = newApplyDefaultService()
	}

	for initState != nil {
		cplogs.V(5).Infof("Handling state %s", initState.Name())
		cplogs.Flush()

		err := initState.Handle()
		if err != nil {
			return err
		}
		initState = initState.Next()
	}
	i.config.Set(config.InitStatus, initStateCompleted)
	i.config.Save()

	apiKey, err := i.config.GetString(config.ApiKey)
	if err != nil {
		return err
	}
	remoteEnvId, err := i.config.GetString(config.RemoteEnvironmentId)
	if err != nil {
		return err
	}
	flowId, err := i.config.GetString(config.FlowId)
	if err != nil {
		return err
	}

	i.api.SetApiKey(apiKey)

	remoteEnv, err := i.api.GetRemoteEnvironmentStatus(flowId, remoteEnvId)
	if err != nil {
		return err
	}

	fmt.Printf("\n\n# Get started !\n")
	fmt.Println("You can now run `cp-remote watch` to watch your local changes with the deployed environment ! Your deployed environment can be found at this address:")
	if len(remoteEnv.PublicEndpoints) >= 0 && remoteEnv.PublicEndpoints[0].Address != "" {
		fmt.Printf("https://%s", remoteEnv.PublicEndpoints[0].Address)
	}
	fmt.Printf("\n\nCheckout the documentation at https://docs.continuouspipe.io/remote-development/ \n")

	return nil
}

type parseSaveTokenInfo struct {
	config config.ConfigProvider
	token  string
	api    cpapi.CpApiProvider
}

func (p parseSaveTokenInfo) Next() initialization.InitState {
	return newTriggerBuild()
}

func (p parseSaveTokenInfo) Name() string {
	return initStateParseSaveToken
}

func (p parseSaveTokenInfo) Handle() error {
	p.config.Set(config.InitStatus, p.Name())
	p.config.Save()

	//we expect the token to have: api-key, remote-environment-id, project, cp-username, git-branch
	splitToken := strings.Split(p.token, ",")
	apiKey := splitToken[0]
	remoteEnvId := splitToken[1]
	flowId := splitToken[2]
	cpUsername := splitToken[3]
	gitBranch := splitToken[4]

	cplogs.V(5).Infof("flowId: %s", flowId)
	cplogs.V(5).Infof("remoteEnvId: %s", remoteEnvId)
	cplogs.V(5).Infof("cpUsername: %s", cpUsername)
	cplogs.V(5).Infof("gitBranch: %s", gitBranch)

	//check the status of the build on CP to determine if we need to force push or not
	p.api.SetApiKey(apiKey)
	cplogs.V(5).Infof("fetching remote environment info for user: %s", cpUsername)
	_, err := p.api.GetRemoteEnvironmentStatus(flowId, remoteEnvId)
	if err != nil {
		cplogs.Flush()
		return err
	}

	cplogs.V(5).Infof("saving parsed token info for user: %s", cpUsername)
	//if there are no errors when fetching the remote environment information we can store the token info
	p.config.Set(config.Username, cpUsername)
	p.config.Set(config.ApiKey, apiKey)
	p.config.Set(config.FlowId, flowId)
	p.config.Set(config.RemoteBranch, gitBranch)
	p.config.Set(config.RemoteEnvironmentId, remoteEnvId)
	p.config.Save()
	cplogs.V(5).Infof("saved parsed token info for user: %s", cpUsername)
	cplogs.Flush()
	return nil
}

type triggerBuild struct {
	config   config.ConfigProvider
	api      cpapi.CpApiProvider
	commit   git.CommitExecutor
	lsRemote git.LsRemoteExecutor
	push     git.PushExecutor
	revParse git.RevParseExecutor
	writer   io.Writer
	qp       util.QuestionPrompter
}

func newTriggerBuild() *triggerBuild {
	return &triggerBuild{
		config.C,
		cpapi.NewCpApi(),
		git.NewCommit(),
		git.NewLsRemote(),
		git.NewPush(),
		git.NewRevParse(),
		os.Stdout,
		util.NewQuestionPrompt(),
	}
}

func (p triggerBuild) Next() initialization.InitState {
	return newWaitEnvironmentReady()
}

func (p triggerBuild) Name() string {
	return initStateTriggerBuild
}

func (p triggerBuild) Handle() error {
	p.config.Set(config.InitStatus, p.Name())
	p.config.Save()

	apiKey, err := p.config.GetString(config.ApiKey)
	if err != nil {
		return err
	}
	remoteEnvId, err := p.config.GetString(config.RemoteEnvironmentId)
	if err != nil {
		return err
	}
	flowId, err := p.config.GetString(config.FlowId)
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

	p.api.SetApiKey(apiKey)
	remoteEnv, err := p.api.GetRemoteEnvironmentStatus(flowId, remoteEnvId)
	if err != nil {
		return err
	}

	environments, err := p.api.GetApiEnvironments(flowId)
	if err != nil {
		return err
	}

	envExists := false
	for _, environment := range environments {
		if environment.Identifier == remoteEnv.ClusterIdentifier {
			envExists = true
		}
	}

	cplogs.V(5).Infof("current remote environment status is %s", remoteEnv.Status)

	//if the environment is already running and exists ask the user if he wants to rebuild
	if remoteEnv.Status == cpapi.RemoteEnvironmentRunning && envExists {

		answer := p.qp.RepeatUntilValid("The remote environment is already running, do you want to rebuild it? (yes/no)",
			func(answer string) (bool, error) {
				switch answer {
				case "yes", "no":
					return true, nil
				default:
					return false, fmt.Errorf("Your answer needs to be either yes or no. Your answer was %s", answer)
				}
			})

		cplogs.V(5).Infof("user aswered %s", answer)

		if answer == "no" {
			return nil
		}
	}

	//if the remote environment is not already building, make sure the remote git branch exists
	//and then trigger a build via api
	if remoteEnv.Status != cpapi.RemoteEnvironmentTideRunning {
		cplogs.V(5).Infof("triggering build for the remote environment, user: %s", cpUsername)

		err := p.pushLocalBranchToRemote(remoteName, gitBranch)
		if err != nil {
			return err
		}
		err = p.api.RemoteEnvironmentBuild(flowId, gitBranch)
		if err != nil {
			return err
		}
		fmt.Fprintf(p.writer, "\n# Environment is building...\n")
	}
	return nil
}

func (p triggerBuild) pushLocalBranchToRemote(remoteName string, gitBranch string) error {
	fmt.Fprintf(p.writer, "# Building your environment by push to the branch `%s`\n", gitBranch)
	lbn, err := p.revParse.GetLocalBranchName()
	cplogs.V(5).Infof("local branch name value is %s", lbn)
	if err != nil {
		return err
	}
	p.push.Push(lbn, remoteName, gitBranch)
	return nil
}

func (p triggerBuild) hasRemote(remoteName string, gitBranch string) (bool, error) {
	list, err := p.lsRemote.GetList(remoteName, gitBranch)
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
	config config.ConfigProvider
	api    cpapi.CpApiProvider
	ticker *time.Ticker
	writer io.Writer
}

func newWaitEnvironmentReady() *waitEnvironmentReady {
	return &waitEnvironmentReady{
		config.C,
		cpapi.NewCpApi(),
		time.NewTicker(time.Second * remoteEnvironmentReadinessProbePeriodSeconds),
		os.Stdout,
	}
}

func (p waitEnvironmentReady) Next() initialization.InitState {
	return newApplyEnvironmentSettings()
}

func (p waitEnvironmentReady) Name() string {
	return initStateWaitEnvironmentReady
}

func (p waitEnvironmentReady) Handle() error {
	p.config.Set(config.InitStatus, p.Name())
	p.config.Save()

	apiKey, err := p.config.GetString(config.ApiKey)
	if err != nil {
		return err
	}
	remoteEnvId, err := p.config.GetString(config.RemoteEnvironmentId)
	if err != nil {
		return err
	}
	flowId, err := p.config.GetString(config.FlowId)
	if err != nil {
		return err
	}
	gitBranch, err := p.config.GetString(config.RemoteBranch)
	if err != nil {
		return err
	}

	p.api.SetApiKey(apiKey)
	var remoteEnv *cpapi.ApiRemoteEnvironmentStatus

	remoteEnv, err = p.api.GetRemoteEnvironmentStatus(flowId, remoteEnvId)
	if err != nil {
		return err
	}

	environments, err := p.api.GetApiEnvironments(flowId)
	if err != nil {
		return err
	}

	envExists := false
	for _, environment := range environments {
		if environment.Identifier == remoteEnv.ClusterIdentifier {
			envExists = true
		}
	}

	if remoteEnv.Status == cpapi.RemoteEnvironmentTideFailed {
		fmt.Fprintln(p.writer, "The build had previously failed, retrying..")
		err := p.api.RemoteEnvironmentBuild(flowId, gitBranch)
		if err != nil {
			return err
		}
	}

	if remoteEnv.Status == cpapi.RemoteEnvironmentRunning && envExists {
		return nil
	}

	fmt.Fprintln(p.writer, "Continuous Pipe will now building your developer environment. You can checkout the logs of your first tide there:")
	fmt.Fprintf(p.writer, "https://ui.continuouspipe.io/project/%s/%s/%s/logs\n", remoteEnv.LastTide.Team.Slug, remoteEnv.LastTide.FlowUuid, remoteEnv.LastTide.Uuid)

	s := spinner.New(spinner.CharSets[34], 100*time.Millisecond)
	s.Prefix = "Waiting for the environment to be ready "
	s.Start()

	//wait until the remote environment has been built
	for t := range p.ticker.C {
		cplogs.V(5).Infoln("environment readiness check at ", t)

		remoteEnv, err = p.api.GetRemoteEnvironmentStatus(flowId, remoteEnvId)
		if err != nil {
			break
		}

		cplogs.V(5).Infof("remote environment status is %s", remoteEnv.Status)
		cplogs.Flush()

		switch remoteEnv.Status {
		case cpapi.RemoteEnvironmentTideRunning:
			cplogs.V(5).Infoln("The remote environment is still building")

		case cpapi.RemoteEnvironmentTideNotStarted:
			cplogs.V(5).Infof("re-trying triggering build for the remote environment")
			cplogs.Flush()
			err = p.api.RemoteEnvironmentBuild(flowId, gitBranch)
			break
		case cpapi.RemoteEnvironmentTideFailed:
			err = fmt.Errorf("remote environment id %s cretion has failed. To see more information about the error go to https://ui.continuouspipe.io/", remoteEnvId)
			break
		case cpapi.RemoteEnvironmentRunning:
			cplogs.V(5).Infoln("The remote environment is running")
			s.Stop()
			return nil
		}

	}

	s.Stop()
	return err
}

type applyEnvironmentSettings struct {
	config             config.ConfigProvider
	api                cpapi.CpApiProvider
	kubeCtlInitializer kubectlapi.KubeCtlInitializer
	writer             io.Writer
}

func newApplyEnvironmentSettings() *applyEnvironmentSettings {
	return &applyEnvironmentSettings{
		config.C,
		cpapi.NewCpApi(),
		kubectlapi.NewKubeCtlInit(),
		os.Stdout,
	}
}

func (p applyEnvironmentSettings) Next() initialization.InitState {
	return newApplyDefaultService()
}

func (p applyEnvironmentSettings) Name() string {
	return initStateApplyEnvironmentSettings
}

func (p applyEnvironmentSettings) Handle() error {
	p.config.Set(config.InitStatus, p.Name())
	p.config.Save()

	apiKey, err := p.config.GetString(config.ApiKey)
	if err != nil {
		return err
	}
	flowId, err := p.config.GetString(config.FlowId)
	if err != nil {
		return err
	}
	remoteEnvId, err := p.config.GetString(config.RemoteEnvironmentId)
	if err != nil {
		return err
	}

	p.api.SetApiKey(apiKey)

	remoteEnv, err := p.api.GetRemoteEnvironmentStatus(flowId, remoteEnvId)
	if err != nil {
		return err
	}

	cplogs.V(5).Infof("saving remote environment info for environment name: %s, environment id: %s", remoteEnv.KubeEnvironmentName, remoteEnvId)
	//the environment has been built, so save locally the settings received from the server
	p.config.Set(config.ClusterIdentifier, remoteEnv.ClusterIdentifier)
	p.config.Set(config.KubeEnvironmentName, remoteEnv.KubeEnvironmentName)
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

func (p applyEnvironmentSettings) applySettingsToCubeCtlConfig() error {
	environment, err := p.config.GetString(config.KubeEnvironmentName)
	if err != nil {
		return err
	}

	err = p.kubeCtlInitializer.Init(environment)
	if err != nil {
		return err
	}
	return nil
}

type applyDefaultService struct {
	config config.ConfigProvider
	qp     util.QuestionPrompter
	ks     services.ServiceFinder
	writer io.Writer
}

func newApplyDefaultService() *applyDefaultService {
	return &applyDefaultService{
		config.C,
		util.NewQuestionPrompt(),
		services.NewKubeService(),
		os.Stdout}
}

func (p applyDefaultService) Next() initialization.InitState {
	return nil
}

func (p applyDefaultService) Name() string {
	return initStateApplyDefaultService
}

func (p applyDefaultService) Handle() error {
	p.config.Set(config.InitStatus, p.Name())
	p.config.Save()

	environment, err := p.config.GetString(config.KubeEnvironmentName)
	if err != nil {
		return err
	}

	list, err := p.ks.FindAll(environment, environment)
	if err != nil {
		return err
	}

	if len(list.Items) == 0 {
		cplogs.V(5).Infoln("No services where found.")
		return nil
	}

	if len(list.Items) == 1 {
		cplogs.V(5).Infoln("Only 1 service found, setting that one as default.")
		p.config.Set(list.Items[0].GetName(), config.Service)
		p.config.Save()
		return nil
	}

	fmt.Fprintln(p.writer, "# Last steps!")

	var options string
	for key, s := range list.Items {
		options = options + fmt.Sprintf("[%d] %s\n", key, s.GetName())
	}
	question := fmt.Sprintf("Which default container would you like to use?\n"+
		"%s\n\n"+
		"Select an option from 0 to %d: ", options, len(list.Items)-1)
	serviceKey := p.qp.RepeatUntilValid(question, func(answer string) (bool, error) {
		for key := range list.Items {
			if strconv.Itoa(key) == answer {
				return true, nil
			}
		}
		return false, fmt.Errorf("Please select an option between [0-%d]", len(list.Items))

	})
	key, err := strconv.Atoi(serviceKey)
	if err != nil {
		return err
	}
	serviceName := list.Items[key].GetName()
	p.config.Set(config.Service, serviceName)
	p.config.Save()
	return nil
}
