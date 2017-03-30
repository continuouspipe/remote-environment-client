package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/benchmark"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	"github.com/continuouspipe/remote-environment-client/sync"
	"github.com/continuouspipe/remote-environment-client/sync/monitor"
	"github.com/continuouspipe/remote-environment-client/sync/options"
	"github.com/continuouspipe/remote-environment-client/util"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var pushSyncExample = `
# push files and folders to the remote pod
%[1]s %[2]s

# push files and folders to the remote pod specifying the environment
%[1]s %[2]s -e techup-dev-user -s web
`

func NewSyncCmd() *cobra.Command {
	pu := NewPushCmd()
	pu.Use = "sync"
	pu.Short = "Sync local changes to the remote filesystem (alias for push)"
	pu.Aliases = []string{"sy"}
	pu.Example = fmt.Sprintf(pushSyncExample, config.AppName, "sync")
	return pu
}

func NewPushCmd() *cobra.Command {
	settings := config.C
	handler := &PushHandle{}
	handler.qp = util.NewQuestionPrompt()
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	handler.writer = os.Stdout

	command := &cobra.Command{
		Use:     "push",
		Aliases: []string{"pu"},
		Short:   "Push local changes to the remote filesystem",
		Example: fmt.Sprintf(pushSyncExample, config.AppName, "push"),
		Long: `The push command will copy changes from the local to the remote filesystem.
Note that this will delete any files/folders in the remote container that are not present locally.`,
		Run: func(cmd *cobra.Command, args []string) {
			validateConfig()

			exclusion := monitor.NewExclusion()
			_, err := exclusion.WriteDefaultExclusionsToFile()
			checkErr(err)

			fmt.Println("Push in progress")

			benchmrk := benchmark.NewCmdBenchmark()
			benchmrk.Start("push")

			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()
			syncer := sync.GetSyncer()

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(args, podsFinder, podsFilter, syncer))

			_, err = benchmrk.StopAndLog()
			checkErr(err)
			cplogs.Flush()
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

type PushHandle struct {
	Command     *cobra.Command
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
func (h *PushHandle) Complete(cmd *cobra.Command, argsIn []string, settings *config.Config) error {
	h.Command = cmd

	var err error
	if h.options.environment == "" {
		h.options.environment, err = settings.GetString(config.KubeEnvironmentName)
		checkErr(err)
	}
	if h.options.service == "" {
		h.options.service, err = settings.GetString(config.Service)
		checkErr(err)
	}
	if strings.HasSuffix(h.options.remoteProjectPath, "/") == false {
		h.options.remoteProjectPath = h.options.remoteProjectPath + "/"
	}
	return nil
}

// Validate checks that the provided push options are specified.
func (h *PushHandle) Validate() error {
	if len(strings.Trim(h.options.environment, " ")) == 0 {
		return fmt.Errorf("the environment specified is invalid")
	}
	if len(strings.Trim(h.options.service, " ")) == 0 {
		return fmt.Errorf("the service specified is invalid")
	}
	if strings.HasPrefix(h.options.remoteProjectPath, "/") == false {
		return fmt.Errorf("please specify an absolute path for your --remote-project-path")
	}
	return nil
}

// Copies all the files and folders from the current directory into the remote container
func (h *PushHandle) Handle(args []string, podsFinder pods.Finder, podsFilter pods.Filter, syncer sync.Syncer) error {
	if h.options.delete {
		if h.options.yall == false {
			answer := deleteFlagWarning(h.qp)
			if answer == "no" {
				return nil
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
		return nil
	}

	allPods, err := podsFinder.FindAll(user, apiKey, addr, h.options.environment)
	if err != nil {
		return err
	}

	pod, err := podsFilter.ByService(allPods, h.options.service)
	if err != nil {
		return err
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
			return err
		}
		paths = append(paths, absFilePath)
	}

	err = syncer.Sync(paths)
	fmt.Fprintf(h.writer, "Push complete, the files and folders that has been sent can be found in the logs %s\n", cplogs.GetLogInfoFile())
	return err
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
