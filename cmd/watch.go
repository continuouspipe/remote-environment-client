package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	remotecplogs "github.com/continuouspipe/remote-environment-client/cplogs/remote"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/session"
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/continuouspipe/remote-environment-client/sync/monitor"
	"github.com/continuouspipe/remote-environment-client/sync/options"
	"github.com/continuouspipe/remote-environment-client/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

//WatchCmdName is the command name identifier
const WatchCmdName = "watch"

func NewWatchCmd() *cobra.Command {
	settings := config.C
	handler := &WatchHandle{}
	handler.qp = util.NewQuestionPrompt()
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	handler.api = cpapi.NewCpAPI()
	handler.config = settings
	handler.writer = os.Stdout

	command := &cobra.Command{
		Use:     WatchCmdName,
		Aliases: []string{"wa"},
		Short:   msgs.WatchCommandShortDescription,
		Long:    msgs.WatchCommandLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			remoteCommand := remotecplogs.NewRemoteCommand(WatchCmdName, os.Args)
			cs := session.NewCommandSession().Start()

			//validate the configuration file
			missingSettings, ok := config.C.Validate()
			if ok == false {
				reason := fmt.Sprintf(msgs.InvalidConfigSettings, missingSettings)
				err := remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.Ended(http.StatusBadRequest, reason, "", *cs))
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cs, err)
				cperrors.ExitWithMessage(reason)
			}

			suggestion, err := RunWatch(handler, args, settings)
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

	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	service, err := settings.GetString(config.Service)
	checkErr(err)

	command.PersistentFlags().StringVarP(&handler.options.environment, config.KubeEnvironmentName, "e", environment, "The full remote environment name")
	command.PersistentFlags().StringVarP(&handler.options.service, config.Service, "s", service, "The service to use (e.g.: web, mysql)")
	command.PersistentFlags().Int64VarP(&handler.options.latency, "latency", "l", 500, "Sync latency / speed in milli-seconds")
	command.PersistentFlags().IntVarP(&handler.options.individualFileSyncThreshold, "individual-file-sync-threshold", "t", 10, "Above this threshold the watch command will sync any file or folder that is different compared to the local one")
	command.PersistentFlags().StringVarP(&handler.options.remoteProjectPath, "remote-project-path", "a", "/app/", "Specify the absolute path to your project folder, by default set to /app/")
	command.PersistentFlags().BoolVar(&handler.options.dryRun, "dry-run", false, "Show what would have been transferred")
	command.PersistentFlags().BoolVar(&handler.options.rsyncVerbose, "rsync-verbose", false, "Allows to use rsync in verbose mode and debug issues with exclusions")
	command.PersistentFlags().BoolVar(&handler.options.delete, "delete", false, "Delete extraneous files from destination directories")
	command.PersistentFlags().BoolVarP(&handler.options.yall, "yes", "y", false, "Skip warning")
	return command
}

func RunWatch(handler *WatchHandle, args []string, settings *config.Config) (reason string, err error) {
	dirMonitor := monitor.GetOsDirectoryMonitor()

	exclusion := monitor.NewExclusion()
	_, err = exclusion.WriteDefaultExclusionsToFile()
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionWriteDefaultExclusionFileFailed, monitor.CustomExclusionsFile, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "write to the default exclusion file before push has failed").String())
	}

	podsFinder := pods.NewKubePodsFind()
	podsFilter := pods.NewKubePodsFilter()

	handler.Stdout = os.Stdout
	handler.syncer = sync.GetSyncer()

	handler.Complete(args, settings)

	suggestion, err := handler.Validate()
	if err != nil {
		return suggestion, err
	}

	suggestion, err = handler.Handle(dirMonitor, podsFinder, podsFilter)
	if err != nil {
		return suggestion, err
	}

	return "", nil
}

type WatchHandle struct {
	Stdout      io.Writer
	syncer      sync.Syncer
	kubeCtlInit kubectlapi.KubeCtlInitializer
	api         cpapi.DataProvider
	config      config.ConfigProvider
	writer      io.Writer
	qp          util.QuestionPrompter
	options     watchCmdOptions
}

type watchCmdOptions struct {
	environment, service, remoteProjectPath string
	latency                                 int64
	individualFileSyncThreshold             int
	rsyncVerbose, dryRun, delete, yall      bool
}

// Complete verifies command line arguments and loads data from the command environment
func (h *WatchHandle) Complete(argsIn []string, settings *config.Config) {
	if h.options.environment == "" {
		h.options.environment = settings.GetStringQ(config.KubeEnvironmentName)
	}
	if h.options.service == "" {
		h.options.service = settings.GetStringQ(config.Service)
	}
	if strings.HasSuffix(h.options.remoteProjectPath, "/") == false {
		h.options.remoteProjectPath = h.options.remoteProjectPath + "/"
	}
}

// Validate checks that the provided watch options are specified.
func (h *WatchHandle) Validate() (suggestion string, err error) {
	if len(strings.Trim(h.options.environment, " ")) == 0 {
		return msgs.EnvironmentSpecifiedEmpty, errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.EnvironmentSpecifiedEmpty).String())
	}
	if len(strings.Trim(h.options.service, " ")) == 0 {
		return msgs.ServiceSpecifiedEmpty, errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.ServiceSpecifiedEmpty).String())
	}
	if h.options.latency <= 100 {
		return msgs.LatencyValueTooSmall, errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.LatencyValueTooSmall).String())
	}
	if strings.HasPrefix(h.options.remoteProjectPath, "/") == false {
		return msgs.RemoteProjectPathEmpty, errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.RemoteProjectPathEmpty).String())
	}
	return "", nil
}

func (h *WatchHandle) Handle(dirMonitor monitor.DirectoryMonitor, podsFinder pods.Finder, podsFilter pods.Filter) (suggestion string, err error) {
	if h.options.delete {
		if h.options.yall == false {
			answer := deleteFlagWarning(h.qp)
			if answer == "no" {
				return "", nil
			}
		}
		fmt.Fprintln(h.writer, "Delete mode enabled.")
	} else {
		fmt.Fprintln(h.writer, "Delete mode disabled. If you need to enable it use the --delete flag")
	}

	if h.options.dryRun {
		fmt.Fprintln(h.writer, "Dry run mode enabled")
	}

	fmt.Fprintf(h.writer, "\nWatching for changes. Quit anytime with Ctrl-C.\n")

	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionGetSettingsError, session.CurrentSession.SessionID), err
	}

	allPods, err := podsFinder.FindAll(user, apiKey, addr, h.options.environment)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionFindPodsFailed, session.CurrentSession.SessionID), err
	}

	pod := podsFilter.List(*allPods).ByService(h.options.service).ByStatus("Running").ByStatusReason("Running").First()
	if pod == nil {
		return fmt.Sprintf(msgs.SuggestionRunningPodNotFound, h.options.service, h.options.environment, config.AppName, "bash", session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, fmt.Sprintf(msgs.NoActivePodsFoundForSpecifiedServiceName, h.options.service)).String())
	}

	remoteEnvId := h.config.GetStringQ(config.RemoteEnvironmentId)
	flowId := h.config.GetStringQ(config.FlowId)

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf(msgs.PleaseContactSupport, session.CurrentSession.SessionID), err
	}

	h.api.SetAPIKey(apiKey)
	remoteEnv, err := h.api.GetRemoteEnvironmentStatus(flowId, remoteEnvId)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionGetEnvironmentStatusFailed, session.CurrentSession.SessionID), err
	}
	cpapi.PrintPublicEndpoints(h.Stdout, remoteEnv.PublicEndpoints)

	syncOptions := options.SyncOptions{}
	syncOptions.KubeConfigKey = h.options.environment
	syncOptions.Environment = h.options.environment
	syncOptions.Pod = pod.GetName()
	syncOptions.IndividualFileSyncThreshold = h.options.individualFileSyncThreshold
	syncOptions.RemoteProjectPath = h.options.remoteProjectPath
	syncOptions.DryRun = h.options.dryRun
	syncOptions.Verbose = h.options.rsyncVerbose
	syncOptions.Delete = h.options.delete
	h.syncer.SetOptions(syncOptions)

	dirMonitor.SetLatency(time.Duration(h.options.latency))

	fmt.Fprintf(h.Stdout, "\nDestination Pod: %s\n", pod.GetName())

	observer := sync.GetSyncOnEventObserver(h.syncer)

	err = dirMonitor.AnyEventCall(cwd, observer)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionDirectoryMonitorFailed, session.CurrentSession.SessionID), err
	}
	return "", nil
}
