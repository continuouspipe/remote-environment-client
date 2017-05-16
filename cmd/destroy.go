package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	remotecplogs "github.com/continuouspipe/remote-environment-client/cplogs/remote"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/git"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/session"
	"github.com/continuouspipe/remote-environment-client/util"
	"github.com/spf13/cobra"
)

//DestroyCmdName is the name identifier for the destroy command
const DestroyCmdName = "destroy"

//NewDestroyCmd return a new cobra command that handles the destruction of a remote environment
func NewDestroyCmd() *cobra.Command {
	handler := NewDestroyHandle()
	handler.api = cpapi.NewCpAPI()
	handler.config = config.C
	handler.stdout = os.Stdout
	handler.lsRemote = git.NewLsRemote()
	handler.push = git.NewPush()
	handler.qp = util.NewQuestionPrompt()
	command := &cobra.Command{
		Use:   DestroyCmdName,
		Short: msgs.DestroyCommandShortDescription,
		Long:  msgs.DestroyCommandLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			remoteCommand := remotecplogs.NewRemoteCommand(DestroyCmdName, args)
			cs := session.NewCommandSession().Start()

			suggestion, err := handler.Handle()
			if err != nil {
				code, reason, stack := cperrors.FindCause(err)
				err = remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.Ended(code, reason, stack, *cs))
				if err != nil {
					cplogs.V(4).Infof(remotecplogs.ErrorFailedToSendDataToLoggingAPI)
					cplogs.Flush()
				}
				cperrors.ExitWithMessage(suggestion)
			}

			err = remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.EndedOk(*cs))
			if err != nil {
				cplogs.V(4).Infof(remotecplogs.ErrorFailedToSendDataToLoggingAPI)
				cplogs.Flush()
			}
		},
	}
	return command
}

//DestroyHandle holds the dependencies of the destroy handler
type DestroyHandle struct {
	api      cpapi.DataProvider
	config   config.ConfigProvider
	lsRemote git.LsRemoteExecutor
	push     git.PushExecutor
	qp       util.QuestionPrompter
	stdout   io.Writer
}

//NewDestroyHandle ctor for the DestroyHandle struct
func NewDestroyHandle() *DestroyHandle {
	return &DestroyHandle{}
}

//Handle cancels a running tide if exists, then deletes the remote environment via CP Api and finally proceed with the deletion of the git remote branch
func (h *DestroyHandle) Handle() (suggestion string, err error) {
	answer := h.qp.RepeatUntilValid(fmt.Sprintf(msgs.ItWillDeleteGitBranchAndRemoteEnvironment, msgs.YesNoOptions),
		func(answer string) (bool, error) {
			switch answer {
			case "yes", "no":
				return true, nil
			default:
				return false, fmt.Errorf(msgs.InvalidAnswerForYesNo, answer)
			}
		})

	if answer == "no" {
		return "", nil
	}

	apiKey := h.config.GetStringQ(config.ApiKey)
	flowID := h.config.GetStringQ(config.FlowId)
	environment := h.config.GetStringQ(config.KubeEnvironmentName)
	remoteEnvironmentID := h.config.GetStringQ(config.RemoteEnvironmentId)
	cluster := h.config.GetStringQ(config.ClusterIdentifier)
	remoteName := h.config.GetStringQ(config.RemoteName)
	gitBranch := h.config.GetStringQ(config.RemoteBranch)

	if apiKey != "" && flowID != "" && remoteEnvironmentID != "" {
		h.api.SetAPIKey(apiKey)
		//stop building any flows associated with the git branch
		h.api.CancelRunningTide(flowID, remoteEnvironmentID)

		if cluster != "" {
			//delete the remote environment via cp api
			err = h.api.RemoteEnvironmentDestroy(flowID, environment, cluster)
			if err != nil {
				return fmt.Sprintf(msgs.SuggestionRemoteEnvironmentDestroyFailed, session.CurrentSession.SessionID), err
			}
		}
	}

	if remoteName != "" && gitBranch != "" {
		//if remote exists delete remote branch
		remoteExists, err := h.hasRemote(remoteName, gitBranch)
		if err != nil {
			return fmt.Sprintf(msgs.SuggestionGitHasRemoteFailed, session.CurrentSession.SessionID, err.Error()), err
		}

		if remoteExists == true {
			_, err = h.push.DeleteRemote(remoteName, gitBranch)
			if err != nil {
				return fmt.Sprintf(msgs.SuggestionGitDeleteHasFailed, session.CurrentSession.SessionID, err.Error()), err
			}
		}
	}

	return "", nil
}

func (h *DestroyHandle) hasRemote(remoteName string, gitBranch string) (bool, error) {
	list, err := h.lsRemote.GetList(remoteName, gitBranch)
	if err != nil {
		return false, err
	}
	if len(list) == 0 {
		return false, err
	}
	return true, err
}
