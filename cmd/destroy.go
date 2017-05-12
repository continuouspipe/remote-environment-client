package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/git"
	"github.com/continuouspipe/remote-environment-client/util"
	"github.com/spf13/cobra"
)

func NewDestroyCmd() *cobra.Command {
	handler := NewDestroyHandle()
	handler.api = cpapi.NewCpAPI()
	handler.config = config.C
	handler.stdout = os.Stdout
	handler.lsRemote = git.NewLsRemote()
	handler.push = git.NewPush()
	handler.qp = util.NewQuestionPrompt()
	command := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy the remote environment",
		Long: `The destroy command will delete the remote branch used for your remote
environment, ContinuousPipe will then remove the environment.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkErr(handler.Handle())
		},
	}
	return command
}

type DestroyHandle struct {
	api      cpapi.DataProvider
	config   config.ConfigProvider
	lsRemote git.LsRemoteExecutor
	push     git.PushExecutor
	qp       util.QuestionPrompter
	stdout   io.Writer
}

func NewDestroyHandle() *DestroyHandle {
	return &DestroyHandle{}
}

func (h *DestroyHandle) Handle() error {

	answer := h.qp.RepeatUntilValid("This will delete the remote git branch and remote environment, do you want to proceed (yes/no)",
		func(answer string) (bool, error) {
			switch answer {
			case "yes", "no":
				return true, nil
			default:
				return false, fmt.Errorf("Your answer needs to be either yes or no. Your answer was %s", answer)
			}
		})

	if answer == "no" {
		return nil
	}

	apiKey, err := h.config.GetString(config.ApiKey)
	if err != nil {
		return err
	}
	flowId, err := h.config.GetString(config.FlowId)
	if err != nil {
		return err
	}
	environment, err := h.config.GetString(config.KubeEnvironmentName)
	if err != nil {
		return err
	}
	remoteEnvironmentId, err := h.config.GetString(config.RemoteEnvironmentId)
	if err != nil {
		return err
	}
	cluster, err := h.config.GetString(config.ClusterIdentifier)
	if err != nil {
		return err
	}
	remoteName, err := h.config.GetString(config.RemoteName)
	if err != nil {
		return err
	}
	gitBranch, err := h.config.GetString(config.RemoteBranch)
	if err != nil {
		return err
	}

	if apiKey != "" && flowId != "" && remoteEnvironmentId != "" {
		h.api.SetAPIKey(apiKey)
		//stop building any flows associated with the git branch
		err = h.api.CancelRunningTide(flowId, remoteEnvironmentId)
		if err != nil {


			//TODO: Wrap the error with a high level explanation and suggestion, see messages.go
			return err
		}

		if cluster != "" {
			//delete the remote environment via cp api
			err = h.api.RemoteEnvironmentDestroy(flowId, environment, cluster)
			if err != nil {


				//TODO: Wrap the error with a high level explanation and suggestion, see messages.go
				return err
			}
		}
	}

	if remoteName != "" && gitBranch != "" {
		//if remote exists delete remote branch
		remoteExists, err := h.hasRemote(remoteName, gitBranch)
		if err != nil {


			//TODO: Wrap the error with a high level explanation and suggestion, see messages.go
			return err
		}

		if remoteExists == true {
			_, err = h.push.DeleteRemote(remoteName, gitBranch)
			if err != nil {


				//TODO: Wrap the error with a high level explanation and suggestion, see messages.go
			}
		}
	}

	//TODO: Change to return nil
	return err
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
