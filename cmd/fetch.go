package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	remotecplogs "github.com/continuouspipe/remote-environment-client/cplogs/remote"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/session"
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/continuouspipe/remote-environment-client/sync/options"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

//FetchCmdName is the command name identifier
const FetchCmdName = "fetch"

func NewFetchCmd() *cobra.Command {
	settings := config.C
	handler := &FetchHandle{}
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	handler.writer = os.Stdout

	command := &cobra.Command{
		Use:     FetchCmdName,
		Aliases: []string{"fe"},
		Short:   msgs.FetchCommandShortDescription,
		Example: fmt.Sprintf(msgs.FetchCommandExampleDescription, config.AppName),
		Long:    msgs.FetchCommandLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			remoteCommand := remotecplogs.NewRemoteCommand(FetchCmdName, os.Args)
			cmdSession := session.NewCommandSession().Start()

			//validate the configuration file
			missingSettings, ok := config.C.Validate()
			if ok == false {
				reason := fmt.Sprintf(msgs.InvalidConfigSettings, missingSettings)
				err := remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.Ended(http.StatusBadRequest, reason, "", *cmdSession))
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cmdSession, err)
				cperrors.ExitWithMessage(reason)
			}

			fmt.Println(msgs.FetchInProgress)

			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()
			fetcher := sync.GetFetcher()

			handler.Complete(cmd, args, settings)

			err := handler.Validate()
			if err != nil {
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cmdSession, err)
				cperrors.ExitWithMessage(err.Error())
			}

			suggestion, err := handler.Handle(args, podsFinder, podsFilter, fetcher)
			if err != nil {
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cmdSession, err)
				cperrors.ExitWithMessage(suggestion)
			}

			err = remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.EndedOk(*cmdSession))
			if err != nil {
				cplogs.V(4).Infof(remotecplogs.ErrorFailedToSendDataToLoggingAPI)
				cplogs.Flush()
			}

			fmt.Println(msgs.FetchCompleted)
			cplogs.Flush()
		},
	}

	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	service, err := settings.GetString(config.Service)
	checkErr(err)

	command.PersistentFlags().StringVarP(&handler.Environment, config.KubeEnvironmentName, "e", environment, "The full remote environment name")
	command.PersistentFlags().StringVarP(&handler.Service, config.Service, "s", service, "The service to use (e.g.: web, mysql)")
	command.PersistentFlags().StringVarP(&handler.File, "file", "f", "", "Allows to specify a file that needs to be fetch from the pod")
	command.PersistentFlags().StringVarP(&handler.RemoteProjectPath, "remote-project-path", "a", "/app/", "Specify the absolute path to your project folder, by default set to /app/")
	command.PersistentFlags().BoolVar(&handler.rsyncVerbose, "rsync-verbose", false, "Allows to use rsync in verbose mode and debug issues with exclusions")
	command.PersistentFlags().BoolVar(&handler.dryRun, "dry-run", false, "Show what would have been transferred")
	return command
}

type FetchHandle struct {
	Command           *cobra.Command
	Environment       string
	Service           string
	File              string
	RemoteProjectPath string
	kubeCtlInit       kubectlapi.KubeCtlInitializer
	rsyncVerbose      bool
	dryRun            bool
	writer            io.Writer
}

// Complete verifies command line arguments and loads data from the command environment
func (h *FetchHandle) Complete(cmd *cobra.Command, argsIn []string, settings *config.Config) {
	h.Command = cmd
	if h.Environment == "" {
		h.Environment = settings.GetStringQ(config.KubeEnvironmentName)
	}
	if h.Service == "" {
		h.Service = settings.GetStringQ(config.Service)
	}
	if strings.HasSuffix(h.RemoteProjectPath, "/") == false {
		h.RemoteProjectPath = h.RemoteProjectPath + "/"
	}
}

// Validate checks that the provided fetch options are specified.
func (h *FetchHandle) Validate() error {
	if len(strings.Trim(h.Environment, " ")) == 0 {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.EnvironmentSpecifiedEmpty).String())
	}
	if len(strings.Trim(h.Service, " ")) == 0 {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.ServiceSpecifiedEmpty).String())
	}
	if strings.HasPrefix(h.RemoteProjectPath, "/") == false {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.RemoteProjectPathEmpty).String())
	}
	return nil
}

// Copies all the files and folders from the remote development environment into the current directory
func (h *FetchHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter, fetcher sync.Fetcher) (suggestion string, err error) {
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionGetSettingsError, session.CurrentSession.SessionID), err
	}

	allPods, err := podsFinder.FindAll(user, apiKey, addr, h.Environment)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionFindPodsFailed, session.CurrentSession.SessionID), err
	}

	pod := podsFilter.List(*allPods).ByService(h.Service).ByStatus("Running").ByStatusReason("Running").First()
	if pod == nil {
		return fmt.Sprintf(msgs.SuggestionRunningPodNotFound, h.Service, h.Environment, config.AppName, "bash", session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, fmt.Sprintf(msgs.NoActivePodsFoundForSpecifiedServiceName, h.Service)).String())
	}

	if h.dryRun {
		fmt.Fprintln(h.writer, "Dry run mode enabled")
	}

	syncOptions := options.SyncOptions{}
	syncOptions.Verbose = h.rsyncVerbose
	syncOptions.Environment = h.Environment
	syncOptions.KubeConfigKey = h.Environment
	syncOptions.Pod = pod.GetName()
	syncOptions.RemoteProjectPath = h.RemoteProjectPath
	syncOptions.DryRun = h.dryRun
	fetcher.SetOptions(syncOptions)
	err = fetcher.Fetch(h.File)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionFetchFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "error while running rsync").String())
	}
	return "", nil
}
