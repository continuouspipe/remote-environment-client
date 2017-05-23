package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	"github.com/continuouspipe/remote-environment-client/sync/monitor"
	"github.com/continuouspipe/remote-environment-client/sync/options"
	"github.com/continuouspipe/remote-environment-client/util"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

//SyncCmdName is the command name identifier
const SyncCmdName = "sync"

//PushCmdName is the command name identifier
const PushCmdName = "push"

func NewSyncCmd() *cobra.Command {
	pu := NewPushCmd()
	pu.Use = SyncCmdName
	pu.Short = msgs.SyncCommandShortDescription
	pu.Long = msgs.SyncCommandLongDescription
	pu.Aliases = []string{"sy"}
	pu.Example = fmt.Sprintf(msgs.PushCmdExampleDescription, config.AppName, "sync")
	return pu
}

func NewPushCmd() *cobra.Command {
	settings := config.C
	handler := &PushHandle{}
	handler.qp = util.NewQuestionPrompt()
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	handler.writer = os.Stdout

	command := &cobra.Command{
		Use:     PushCmdName,
		Aliases: []string{"pu"},
		Short:   msgs.PushCommandShortDescription,
		Example: fmt.Sprintf(msgs.PushCmdExampleDescription, config.AppName, "push"),
		Long:    msgs.PushCommandLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			remoteCommand := remotecplogs.NewRemoteCommand(PushCmdName, os.Args)
			cs := session.NewCommandSession().Start()

			//validate the configuration file
			missingSettings, ok := config.C.Validate()
			if ok == false {
				reason := fmt.Sprintf(msgs.InvalidConfigSettings, missingSettings)
				err := remotecplogs.NewRemoteCommandSender().Send(*remoteCommand.Ended(http.StatusBadRequest, reason, "", *cs))
				remotecplogs.EndSessionAndSendErrorCause(remoteCommand, cs, err)
				cperrors.ExitWithMessage(reason)
			}

			suggestion, err := RunPush(handler, args, settings)
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
	command.PersistentFlags().StringVarP(&handler.options.file, "file", "f", "", "Allows to specify a file that needs to be pushed to the pod")
	command.PersistentFlags().StringVarP(&handler.options.remoteProjectPath, "remote-project-path", "a", "/app/", "Specify the absolute path to your project folder, by default set to /app/")
	command.PersistentFlags().BoolVar(&handler.options.rsyncVerbose, "rsync-verbose", false, "Allows to use rsync in verbose mode and debug issues with exclusions")
	command.PersistentFlags().BoolVar(&handler.options.dryRun, "dry-run", false, "Show what would have been transferred")
	command.PersistentFlags().BoolVar(&handler.options.delete, "delete", false, "Delete extraneous files from destination directories")
	command.PersistentFlags().BoolVarP(&handler.options.yall, "yes", "y", false, "Skip warning")

	return command
}

func RunPush(handler *PushHandle, args []string, settings *config.Config) (suggestion string, err error) {
	exclusion := monitor.NewExclusion()
	_, err = exclusion.WriteDefaultExclusionsToFile()
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionWriteDefaultExclusionFileFailed, monitor.CustomExclusionsFile, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "write to the default exclusion file before push has failed").String())
	}

	fmt.Fprintln(handler.writer, msgs.PushInProgress)

	podsFinder := pods.NewKubePodsFind()
	podsFilter := pods.NewKubePodsFilter()
	syncer := sync.GetSyncer()

	handler.Complete(args, settings)

	err = handler.Validate()
	if err != nil {
		return err.Error(), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, "error when validating the options before push").String())
	}

	suggestion, err = handler.Handle(args, podsFinder, podsFilter, syncer)
	if err != nil {
		return suggestion, err
	}
	return "", nil
}

type PushHandle struct {
	kubeCtlInit kubectlapi.KubeCtlInitializer
	writer      io.Writer
	qp          util.QuestionPrompter
	options     pushCmdOptions
}

type pushCmdOptions struct {
	environment, service, remoteProjectPath, file string
	rsyncVerbose, dryRun, delete, yall            bool
}

// Complete verifies command line arguments and loads data from the command environment
func (h *PushHandle) Complete(argsIn []string, settings *config.Config) {
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

// Validate checks that the provided push options are specified.
func (h *PushHandle) Validate() error {
	if len(strings.Trim(h.options.environment, " ")) == 0 {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.EnvironmentSpecifiedEmpty).String())
	}
	if len(strings.Trim(h.options.service, " ")) == 0 {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.ServiceSpecifiedEmpty).String())
	}
	if strings.HasPrefix(h.options.remoteProjectPath, "/") == false {
		return errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.RemoteProjectPathEmpty).String())
	}
	return nil
}

// Copies all the files and folders from the current directory into the remote container
func (h *PushHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter, syncer sync.Syncer) (suggestion string, err error) {
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
		return fmt.Sprintf(msgs.SuggestionRunningPodNotFound, h.options.service, h.options.environment, config.AppName, PushCmdName, session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, fmt.Sprintf(msgs.NoActivePodsFoundForSpecifiedServiceName, h.options.service)).String())
	}

	syncOptions := options.SyncOptions{}
	//set individual file threshold to 1 as for now we only allow the user to specify 1 file to be pushed
	syncOptions.IndividualFileSyncThreshold = 1
	syncOptions.Verbose = h.options.rsyncVerbose
	syncOptions.Environment = h.options.environment
	syncOptions.KubeConfigKey = h.options.environment
	syncOptions.Pod = pod.GetName()
	syncOptions.RemoteProjectPath = h.options.remoteProjectPath
	syncOptions.DryRun = h.options.dryRun
	syncOptions.Delete = h.options.delete
	syncer.SetOptions(syncOptions)

	var paths []string
	if h.options.file != "" {
		absFilePath, err := filepath.Abs(h.options.file)
		if err != nil {
			return fmt.Sprintf(msgs.SuggestionFailedToDetermineTheAbsPath, h.options.file, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, fmt.Sprintf("error when taking the absolute path for the file %s", h.options.file)).String())
		}
		paths = append(paths, absFilePath)
	}

	err = syncer.Sync(paths)
	if err != nil {
		return fmt.Sprintf(msgs.SuggestionPushFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, "error while running rsync").String())
	}
	fmt.Fprintf(h.writer, "Push complete, the files and folders that has been sent can be found in the logs %s\n", cplogs.GetLogInfoFile())
	return "", nil
}

func deleteFlagWarning(qp util.QuestionPrompter) string {
	suggestedCmd := color.GreenString(`%s push --delete -y --dry-run | grep "deleting"`, config.AppName)
	return qp.RepeatUntilValid(
		"Using the --delete flag will delete any files or folders from the remote pod that are not found locally.\n"+
			"If you wish to preserve any remote files or folders that are not found locally you can include them in the .cp-remote-ignore file.\n"+
			fmt.Sprintf("If you are unsure about what files will potentially be deleted you can run %s to find out.\n", suggestedCmd)+
			"\nDo you want to proceed (yes/no): ",
		func(answer string) (bool, error) {
			switch answer {
			case "yes", "no":
				return true, nil
			default:
				return false, fmt.Errorf("Your answer needs to be either yes or no. Your answer was %s", answer)
			}
		})
}
