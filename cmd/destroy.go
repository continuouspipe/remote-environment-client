package cmd

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/git"
	"github.com/spf13/cobra"
	"io"
	"os"
)

func NewDestroyCmd() *cobra.Command {
	handler := &DestroyHandle{}
	handler.api = cpapi.NewCpApi()
	handler.config = config.C
	handler.stdout = os.Stdout
	command := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy the remote environment",
		Long: `The destroy command will delete the remote branch used for your remote
environment, ContinuousPipe will then remove the environment.`,
		Run: func(cmd *cobra.Command, args []string) {
			validateConfig()

			checkErr(handler.Handle())

		},
	}
	return command
}

type DestroyHandle struct {
	api    cpapi.CpApiProvider
	config config.ConfigProvider
	stdout io.Writer
}

func (h *DestroyHandle) Handle() error {
	apiKey, err := h.config.GetString(config.ApiKey)
	if err != nil {
		return err
	}
	remoteEnvId, err := h.config.GetString(config.RemoteEnvironmentId)
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

	h.api.SetApiKey(apiKey)

	remoteEnv, err := h.api.GetRemoteEnvironmentStatus(flowId, remoteEnvId)
	if err != nil {
		return err
	}

	//stop the tide if the remote environment build is running
	if remoteEnv.Status == cpapi.RemoteEnvironmentStatusBuilding {

	}
	//delete the remote environment via cp api
	h.api.RemoteEnvironmentDestroy(flowId, environment, cluster)

	//delete the branch
	push := git.NewPush()
	_, err = push.DeleteRemote(remoteName, gitBranch)
	return err
}
