package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"net/http"

	"github.com/briandowns/spinner"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	remotecplogs "github.com/continuouspipe/remote-environment-client/cplogs/remote"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/git"
	"github.com/continuouspipe/remote-environment-client/initialization"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/services"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/session"
	"github.com/continuouspipe/remote-environment-client/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const initStateParseSaveToken = "parse-save-token"
const initStateTriggerBuild = "trigger-build"
const initStateWaitEnvironmentReady = "wait-environment-ready"
const initStateApplyEnvironmentSettings = "apply-environment-settings"
const initStateApplyDefaultService = "apply-default-service"
const initStateCompleted = "completed"

const remoteEnvironmentReadinessProbePeriodSeconds = 30

//InitCmdName is the name identifier for the build command
const InitCmdName = "init"

//NewInitCmd Initialises the remote environment
func NewInitCmd() *cobra.Command {
	settings := config.C

	command := &cobra.Command{
		Use:     InitCmdName,
		Aliases: []string{"in", "setup"},
		Short:   msgs.InitCommandShortDescription,
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			remoteCommand := remotecplogs.NewRemoteCommand(InitCmdName, os.Args)
			cs := session.NewCommandSession().Start()

			//Mock base64 token when 5 arguments are passed in
			if len(args) == 5 {
				args = []string{base64.StdEncoding.EncodeToString([]byte(strings.Join(args, ",")))}
			}

			interactive, _ := cmd.PersistentFlags().GetBool("interactive")
			reset, _ := cmd.PersistentFlags().GetBool("reset")

			var handler InitStrategy

			if interactive {
				handler = NewInitInteractiveHandler(reset)
			} else {
				remoteName, _ := cmd.PersistentFlags().GetString(config.RemoteName)
				handler = NewInitHandler(remoteName, reset)
			}

			suggestion, err := handler.Complete(args)
			if err != nil {
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cs, err)
				cperrors.ExitWithMessage(suggestion)
			}
			err = handler.Validate()
			if err != nil {
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cs, err)
				cperrors.ExitWithMessage(err.Error())
			}
			suggestion, err = handler.Handle()
			if err != nil {
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cs, err)
				cperrors.ExitWithMessage(suggestion)
			}

			err = remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.EndedOk(*cs))
			if err != nil {
				cplogs.V(4).Infof(remotecplogs.ErrorFailedToSendDataToLoggingAPI)
				cplogs.Flush()
			}
		},
	}

	remoteName, err := settings.GetString(config.RemoteName)
	checkErr(err)
	command.PersistentFlags().String(config.RemoteName, remoteName, "Override the default remote name (origin)")
	command.PersistentFlags().BoolP("reset", "r", false, "With the reset flag set, init will start any partial initializations from the beginning.")
	command.PersistentFlags().BoolP("interactive", "i", false, "Interactive mode allow you specify your cp username and api-key without a token so they can be used with commands that allow the interactive mode.")
	return command
}

//InitStrategy is the interface for the init strategy struct
type InitStrategy interface {
	Complete(argsIn []string) (suggestion string, err error)
	Validate() error
	Handle() (suggestion string, err error)
	SetWriter(io.Writer)
}

//InitInteractiveHandler is a interactive mode strategy where we request the user to insert the global configuration data
//which is the cp username and cp api-key which are mandatory to use the interactive mode on all command that support it.
type InitInteractiveHandler struct {
	config config.ConfigProvider
	qp     util.QuestionPrompter
	api    cpapi.DataProvider
	writer io.Writer
	reset  bool
}

//NewInitInteractiveHandler ctor for InitInteractiveHandler
func NewInitInteractiveHandler(reset bool) *InitInteractiveHandler {
	p := &InitInteractiveHandler{}
	p.api = cpapi.NewCpAPI()
	p.config = config.C
	p.qp = util.NewQuestionPrompt()
	p.reset = reset
	p.writer = os.Stdout
	return p
}

//SetWriter sets the default writer for the struct
func (i *InitInteractiveHandler) SetWriter(writer io.Writer) {
	i.writer = writer
}

//Complete verifies command line arguments and loads data from the command environment
func (i *InitInteractiveHandler) Complete(argsIn []string) (suggestion string, err error) {
	return "", nil
}

//Validate checks that the handler has all it needs
func (i InitInteractiveHandler) Validate() error {
	return nil
}

//Handle request the user the config data required for the interactive mode to work
func (i InitInteractiveHandler) Handle() (suggestion string, err error) {
	username := i.config.GetStringQ(config.Username)
	apiKey := i.config.GetStringQ(config.ApiKey)

	changed := false

	if username == "" || i.reset == true {
		username = i.qp.RepeatIfEmpty("Insert your CP Username:")
		i.config.Set(config.Username, username)
		changed = true

	}
	if apiKey == "" || i.reset == true {
		apiKey = i.qp.RepeatPasswordIfEmpty("Insert your CP Api Key:")
		i.config.Set(config.ApiKey, apiKey)
		changed = true
	}

	if changed == true {
		i.api.SetAPIKey(apiKey)
		user, err := i.api.GetAPIUser(username)
		if err != nil || user.Username != username {
			return fmt.Sprintf(msgs.SuggestionGetAPIUserFailed, username, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "cannot proceed with the initialisation without a valid cp user").String())
		}
		err = i.config.Save(config.GlobalConfigType)
		if err != nil {
			return fmt.Sprintf(msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "failed to save the configuration file while initializing").String())
		}
	}

	fmt.Fprintf(i.writer, "\n# Get started !\n")
	fmt.Fprintf(i.writer, "You can now run commands in interactive mode such as\n%s\n", bashInteractiveFullExample)

	return "", nil
}

//InitHandler is the default initialisation mode which prepare the remote environment so that can be used with any command.
type InitHandler struct {
	interactive bool
	config      config.ConfigProvider
	token       string
	remoteName  string
	reset       bool
	qp          util.QuestionPrompter
	api         cpapi.DataProvider
	writer      io.Writer
}

//NewInitHandler ctor for InitHandler
func NewInitHandler(remoteName string, reset bool) *InitHandler {
	p := &InitHandler{}
	p.api = cpapi.NewCpAPI()
	p.config = config.C
	p.qp = util.NewQuestionPrompt()
	p.remoteName = remoteName
	p.reset = reset
	p.writer = os.Stdout
	return p
}

//SetWriter sets the default writer for the struct
func (i *InitHandler) SetWriter(writer io.Writer) {
	i.writer = writer
}

// Complete verifies command line arguments and loads data from the command environment
func (i *InitHandler) Complete(argsIn []string) (suggestion string, err error) {
	inputToken := ""
	if len(argsIn) > 0 && argsIn[0] != "" {
		inputToken = argsIn[0]
	}

	if inputToken == "" {
		//no given token, attempt to use the one saved in the configuration
		if savedToken, _ := i.config.GetString(config.InitToken); savedToken != "" {
			inputToken = savedToken
		}
	}

	if inputToken != "" {
		decodedToken, err := base64.StdEncoding.DecodeString(inputToken)
		if err != nil {
			return msgs.SuggestionMalformedToken, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "user has provided a malformed token").String())
		}
		i.token = string(decodedToken)
		i.config.Set(config.InitToken, inputToken)
		err = i.config.Save(config.AllConfigTypes)
		if err != nil {
			return fmt.Sprintf(msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "failed to save the configuration file while initializing").String())
		}
	} else {
		return msgs.SuggestionMalformedToken, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "user has provided a malformed token").String())
	}

	if i.remoteName == "" {
		i.remoteName = i.config.GetStringQ(config.RemoteName)
	}

	i.config.Set(config.RemoteName, i.remoteName)
	err = i.config.Save(config.AllConfigTypes)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "failed to save the configuration file while initializing").String())
	}

	return "", nil
}

// Validate checks that the token provided has at least 5 values comma separated
func (i InitHandler) Validate() error {
	splitToken := strings.Split(string(i.token), ",")
	if len(splitToken) != 5 {
		cplogs.V(5).Infof("Token provided %s has %d parts, expected 4", splitToken, len(splitToken))
		return fmt.Errorf("Malformed token. Please go to https://continuouspipe.io/ to obtain a valid token")
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
			return fmt.Errorf("Malformed token. Please go to https://continuouspipe.io/ to obtain a valid token")
		}
	}

	return nil
}

//Handle Executes the initialization
func (i InitHandler) Handle() (suggestion string, err error) {
	currentStatus := i.config.GetStringQ(config.InitStatus)

	if currentStatus == initStateCompleted && i.reset == false {
		answer := i.qp.RepeatIfEmpty("The configuration file is already present, do you want override it and re-initialize? (yes/no)")
		if answer == "no" {
			return "", nil
		}
		cplogs.V(5).Infoln("The user requested to re-initialize the remote environment")
		cplogs.Flush()
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
		initState = &parseSaveTokenInfo{i.config, i.token, cpapi.NewCpAPI()}
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

		suggestion, err := initState.Handle()
		if err != nil {
			return suggestion, err
		}
		initState = initState.Next()
	}
	i.config.Set(config.InitStatus, initStateCompleted)
	err = i.config.Save(config.AllConfigTypes)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID), err

	}

	apiKey := i.config.GetStringQ(config.ApiKey)
	remoteEnvID := i.config.GetStringQ(config.RemoteEnvironmentId)
	flowID := i.config.GetStringQ(config.FlowId)

	i.api.SetAPIKey(apiKey)

	remoteEnv, err := i.api.GetRemoteEnvironmentStatus(flowID, remoteEnvID)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionGetEnvironmentStatusFailed, session.CurrentSession.SessionID), err
	}

	fmt.Fprintf(i.writer, "\n\n# Get started !\n")
	fmt.Fprintln(i.writer, "You can now run `cp-remote watch` to watch your local changes with the deployed environment ! Your deployed environment can be found at this address:")
	cpapi.PrintPublicEndpoints(i.writer, remoteEnv.PublicEndpoints)
	fmt.Fprintf(i.writer, "\n\nCheckout the documentation at https://docs.continuouspipe.io/remote-development/ \n")

	return "", nil
}

type parseSaveTokenInfo struct {
	config config.ConfigProvider
	token  string
	api    cpapi.DataProvider
}

func (p parseSaveTokenInfo) Next() initialization.InitState {
	return newTriggerBuild()
}

func (p parseSaveTokenInfo) Name() string {
	return initStateParseSaveToken
}

func (p parseSaveTokenInfo) Handle() (suggestion string, err error) {
	p.config.Set(config.InitStatus, p.Name())
	err = p.config.Save(config.AllConfigTypes)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID), err
	}
	//we expect the token to have: api-key, remote-environment-id, project, cp-username, git-branch
	splitToken := strings.Split(p.token, ",")
	apiKey := splitToken[0]
	remoteEnvID := splitToken[1]
	flowID := splitToken[2]
	cpUsername := splitToken[3]
	gitBranch := splitToken[4]

	cplogs.V(5).Infof("flowID: %s", flowID)
	cplogs.V(5).Infof("remoteEnvID: %s", remoteEnvID)
	cplogs.V(5).Infof("cpUsername: %s", cpUsername)
	cplogs.V(5).Infof("gitBranch: %s", gitBranch)

	//check the status of the build on CP to determine if we need to force push or not
	p.api.SetAPIKey(apiKey)
	cplogs.V(5).Infof("fetching remote environment info for user: %s", cpUsername)
	_, err = p.api.GetRemoteEnvironmentStatus(flowID, remoteEnvID)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionGetEnvironmentStatusFailed, session.CurrentSession.SessionID), err
	}

	cplogs.V(5).Infof("saving parsed token info for user: %s", cpUsername)
	//if there are no errors when fetching the remote environment information we can store the token info
	p.config.Set(config.Username, cpUsername)
	p.config.Set(config.ApiKey, apiKey)
	p.config.Set(config.FlowId, flowID)
	p.config.Set(config.RemoteBranch, gitBranch)
	p.config.Set(config.RemoteEnvironmentId, remoteEnvID)
	err = p.config.Save(config.AllConfigTypes)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID), err
	}
	cplogs.V(5).Infof("saved parsed token info for user: %s", cpUsername)
	cplogs.Flush()
	return "", nil
}

type triggerBuild struct {
	config   config.ConfigProvider
	api      cpapi.DataProvider
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
		cpapi.NewCpAPI(),
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

func (p triggerBuild) Handle() (suggestion string, err error) {
	p.config.Set(config.InitStatus, p.Name())
	err = p.config.Save(config.AllConfigTypes)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID), err

	}
	apiKey := p.config.GetStringQ(config.ApiKey)
	remoteEnvID := p.config.GetStringQ(config.RemoteEnvironmentId)
	flowID := p.config.GetStringQ(config.FlowId)
	cpUsername := p.config.GetStringQ(config.Username)
	remoteName := p.config.GetStringQ(config.RemoteName)
	gitBranch := p.config.GetStringQ(config.RemoteBranch)

	p.api.SetAPIKey(apiKey)
	remoteEnv, el := p.api.GetRemoteEnvironmentStatus(flowID, remoteEnvID)
	if el != nil {
		return fmt.Sprintf(msgs.SuggestionGetEnvironmentStatusFailed, session.CurrentSession.SessionID), err
	}

	envExists, elr := p.api.RemoteEnvironmentRunningAndExists(flowID, remoteEnvID)
	if elr != nil {
		return fmt.Sprintf(msgs.SuggestionRemoteEnvironmentRunningAndExistsFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "failed to build the remote environment").String())
	}

	cplogs.V(5).Infof("current remote environment status is %s", remoteEnv.Status)
	cplogs.Flush()

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
		cplogs.Flush()
		if answer == "no" {
			return "", nil
		}
	}

	//if the remote environment is not already building, make sure the remote git branch exists
	//and then trigger a build via api
	if remoteEnv.Status != cpapi.RemoteEnvironmentTideRunning {
		cplogs.V(5).Infof("triggering build for the remote environment, user: %s", cpUsername)
		cplogs.Flush()
		err := p.pushLocalBranchToRemote(remoteName, gitBranch)
		if err != nil {
			return fmt.Sprintf(msgs.SuggestionGitPushHasFailed, session.CurrentSession.SessionID, err.Error()), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "we could not trigger the build as git push to the remote branch has failed").String())
		}
		err = p.api.RemoteEnvironmentBuild(flowID, gitBranch)
		if err != nil {
			return fmt.Sprintf(msgs.SuggestionTriggerBuildFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "the trigger build has failed as we could not build the remote environment").String())
		}
		fmt.Fprintf(p.writer, "\n# Environment is building...\n")
	}
	return "", nil
}

func (p triggerBuild) pushLocalBranchToRemote(remoteName string, gitBranch string) error {
	fmt.Fprintf(p.writer, "# Building your environment by push to the branch `%s`\n", gitBranch)
	lbn, err := p.revParse.GetLocalBranchName()
	cplogs.V(5).Infof("local branch name value is %s", lbn)
	cplogs.Flush()
	if err != nil {
		return err
	}
	_, err = p.push.Push(lbn, remoteName, gitBranch)
	return err
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
	api    cpapi.DataProvider
	ticker *time.Ticker
	writer io.Writer
}

func newWaitEnvironmentReady() *waitEnvironmentReady {
	return &waitEnvironmentReady{
		config.C,
		cpapi.NewCpAPI(),
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

func (p waitEnvironmentReady) Handle() (suggestion string, err error) {
	p.config.Set(config.InitStatus, p.Name())
	err = p.config.Save(config.AllConfigTypes)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID), err
	}
	apiKey := p.config.GetStringQ(config.ApiKey)
	remoteEnvID := p.config.GetStringQ(config.RemoteEnvironmentId)
	flowID := p.config.GetStringQ(config.FlowId)
	gitBranch := p.config.GetStringQ(config.RemoteBranch)

	p.api.SetAPIKey(apiKey)
	var remoteEnv *cpapi.APIRemoteEnvironmentStatus

	remoteEnv, el := p.api.GetRemoteEnvironmentStatus(flowID, remoteEnvID)
	if el != nil {
		return fmt.Sprintf(msgs.SuggestionGetEnvironmentStatusFailed, session.CurrentSession.SessionID), err
	}

	envExists, elr := p.api.RemoteEnvironmentRunningAndExists(flowID, remoteEnvID)
	if elr != nil {
		return fmt.Sprintf(msgs.SuggestionRemoteEnvironmentRunningAndExistsFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "failed to build the remote environment").String())

	}

	if (remoteEnv.Status == cpapi.RemoteEnvironmentTideFailed) || (remoteEnv.Status == cpapi.RemoteEnvironmentRunning && envExists == false) {
		fmt.Fprintln(p.writer, "The build had previously failed, retrying..")
		err := p.api.RemoteEnvironmentBuild(flowID, gitBranch)
		if err != nil {
			return fmt.Sprintf(msgs.SuggestionTriggerBuildFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "wait for environment ready has failed as we could not build the remote environment").String())
		}
	}

	fmt.Fprintln(p.writer, "ContinuousPipe is now building your developer environment. You can view the logs of your first tide here:")
	fmt.Fprintf(p.writer, "https://ui.continuouspipe.io/project/%s/%s/%s/logs\n", remoteEnv.LastTide.Team.Slug, remoteEnv.LastTide.FlowUUID, remoteEnv.LastTide.UUID)

	s := spinner.New(spinner.CharSets[34], 100*time.Millisecond)
	s.Prefix = "Waiting for the environment to be ready "
	s.Start()

WAIT_LOOP:
	//wait until the remote environment has been built
	for t := range p.ticker.C {
		cplogs.V(5).Infoln("environment readiness check at ", t)

		remoteEnv, el = p.api.GetRemoteEnvironmentStatus(flowID, remoteEnvID)
		if el != nil {
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
			err = p.api.RemoteEnvironmentBuild(flowID, gitBranch)
			break

		case cpapi.RemoteEnvironmentTideFailed:
			err = fmt.Errorf("remote environment id %s creation has failed. To see more information about the error go to https://ui.continuouspipe.io/", remoteEnvID)
			break WAIT_LOOP

		case cpapi.RemoteEnvironmentRunning:
			cplogs.V(5).Infoln("The remote environment is running")
			//clear any error
			err = nil
			break WAIT_LOOP
		}
	}

	//if there has been an error return it
	if err != nil {
		s.Stop()
		return fmt.Sprintf(msgs.SuggestionTriggerBuildFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "wait for environment ready has failed as we could not build the remote environment").String())
	}

	//when there has been no errors reported, check if the environment actually exist, if not return an error.
	//try 3 times to get a matching environment in case the environment has been created by CP Api is not showing it immediately
	attempts := 3

	envCreated := false
	for i := 0; i < attempts; i++ {
		envCreated, elr = p.api.RemoteEnvironmentRunningAndExists(flowID, remoteEnvID)
		if elr != nil {
			s.Stop()
			return fmt.Sprintf(msgs.SuggestionRemoteEnvironmentRunningAndExistsFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "failed to build the remote environment").String())
		}

		if envCreated {
			cplogs.V(5).Infof("The environment exists")
			break
		}

		cplogs.V(5).Infof("Then environment could not be found, retrying...")
		//sleep for 3 seconds and try again
		time.Sleep(time.Second * remoteEnvironmentReadinessProbePeriodSeconds)
	}

	if !envCreated {
		cplogs.V(5).Infof("The environment could not be found, returning an error")
		err = fmt.Errorf(
			"\nContinuousPipe could not build your developer environment. Please check the logs of your first tide here:\n"+
				"https://ui.continuouspipe.io/project/%s/%s/%s/logs\n"+
				"If there are any changes required to the continuous-pipe.yml file, push them to the repository and retry with cp-remote init [token] --reset.\n", remoteEnv.LastTide.Team.Slug, remoteEnv.LastTide.FlowUUID, remoteEnv.LastTide.UUID)
	}

	s.Stop()
	return "", err
}

type applyEnvironmentSettings struct {
	config             config.ConfigProvider
	api                cpapi.DataProvider
	kubeCtlInitializer kubectlapi.KubeCtlInitializer
	writer             io.Writer
}

func newApplyEnvironmentSettings() *applyEnvironmentSettings {
	return &applyEnvironmentSettings{
		config.C,
		cpapi.NewCpAPI(),
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

func (p applyEnvironmentSettings) Handle() (suggestion string, err error) {
	p.config.Set(config.InitStatus, p.Name())
	err = p.config.Save(config.AllConfigTypes)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID), err
	}

	apiKey := p.config.GetStringQ(config.ApiKey)
	flowID := p.config.GetStringQ(config.FlowId)
	remoteEnvID := p.config.GetStringQ(config.RemoteEnvironmentId)

	p.api.SetAPIKey(apiKey)

	remoteEnv, el := p.api.GetRemoteEnvironmentStatus(flowID, remoteEnvID)
	if el != nil {
		return fmt.Sprintf(msgs.SuggestionGetEnvironmentStatusFailed, session.CurrentSession.SessionID), err
	}

	cplogs.V(5).Infof("saving remote environment info for environment name: %s, environment id: %s", remoteEnv.KubeEnvironmentName, remoteEnvID)
	//the environment has been built, so save locally the settings received from the server
	p.config.Set(config.ClusterIdentifier, remoteEnv.ClusterIdentifier)
	p.config.Set(config.KubeEnvironmentName, remoteEnv.KubeEnvironmentName)
	err = p.config.Save(config.AllConfigTypes)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID), err

	}
	cplogs.V(5).Infoln("saved remote environment info")
	cplogs.Flush()

	err = p.applySettingsToCubeCtlConfig()
	if err != nil {
		return fmt.Sprintf(msgs.PleaseContactSupport, session.CurrentSession.SessionID), err
	}
	return "", nil
}

func (p applyEnvironmentSettings) applySettingsToCubeCtlConfig() error {
	environment, err := p.config.GetString(config.KubeEnvironmentName)
	if err != nil {
		return err
	}

	err = p.kubeCtlInitializer.Init(environment)
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "failed to save the kubectl config file").String())
	}
	return nil
}

type applyDefaultService struct {
	config             config.ConfigProvider
	qp                 util.QuestionPrompter
	ks                 services.ServiceFinder
	kubeCtlInitializer kubectlapi.KubeCtlInitializer
	writer             io.Writer
}

func newApplyDefaultService() *applyDefaultService {
	return &applyDefaultService{
		config.C,
		util.NewQuestionPrompt(),
		services.NewKubeService(),
		kubectlapi.NewKubeCtlInit(),
		os.Stdout}
}

func (p applyDefaultService) Next() initialization.InitState {
	return nil
}

func (p applyDefaultService) Name() string {
	return initStateApplyDefaultService
}

func (p applyDefaultService) Handle() (suggestion string, err error) {
	p.config.Set(config.InitStatus, p.Name())
	err = p.config.Save(config.AllConfigTypes)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID), err

	}
	address, username, apiKey, err := p.kubeCtlInitializer.GetSettings()
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionGetSettingsError, session.CurrentSession.SessionID), err
	}

	environment := p.config.GetStringQ(config.KubeEnvironmentName)

	list, err := p.ks.FindAll(username, apiKey, address, environment)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionFindPodsFailed, session.CurrentSession.SessionID), err
	}

	if len(list.Items) == 0 {
		cplogs.V(5).Infoln("No services where found.")
		return "", nil
	}

	if len(list.Items) == 1 {
		cplogs.V(5).Infoln("Only 1 service found, setting that one as default.")
		p.config.Set(list.Items[0].GetName(), config.Service)
		err := p.config.Save(config.AllConfigTypes)
		if err != nil {
			return fmt.Sprintf(msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID), err
		}
		return "", nil
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
		return fmt.Sprintf(msgs.PleaseContactSupport, session.CurrentSession.SessionID), err
	}
	serviceName := list.Items[key].GetName()
	p.config.Set(config.Service, serviceName)
	err = p.config.Save(config.AllConfigTypes)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionConfigurationSaveFailed, session.CurrentSession.SessionID), err
	}
	return "", nil
}
