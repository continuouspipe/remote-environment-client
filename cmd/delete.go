package cmd

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/spf13/cobra"
	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	kubectlcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

var (
	deleteLong = templates.LongDesc(`
		Delete resources by filenames, stdin, resources and names, or by resources and label selector.

		JSON and YAML formats are accepted. Only one type of the arguments may be specified: filenames,
		resources and names, or resources and label selector.

		Some resources, such as pods, support graceful deletion. These resources define a default period
		before they are forcibly terminated (the grace period) but you may override that value with
		the --grace-period flag, or pass --now to set a grace-period of 1. Because these resources often
		represent entities in the cluster, deletion may not be acknowledged immediately. If the node
		hosting a pod is down or cannot reach the API server, termination may take significantly longer
		than the grace period. To force delete a resource,	you must pass a grace	period of 0 and specify
		the --force flag.

		IMPORTANT: Force deleting pods does not wait for confirmation that the pod's processes have been
		terminated, which can leave those processes running until the node detects the deletion and
		completes graceful deletion. If your processes use shared storage or talk to a remote API and
		depend on the name of the pod to identify themselves, force deleting those pods may result in
		multiple processes running on different machines using the same identification which may lead
		to data corruption or inconsistency. Only force delete pods when you are sure the pod is
		terminated, or if your application can tolerate multiple copies of the same pod running at once.
		Also, if you force delete pods the scheduler may place new pods on those nodes before the node
		has released those resources and causing those pods to be evicted immediately.

		Note that the delete command does NOT do resource version checks, so if someone
		submits an update to a resource right when you submit a delete, their update
		will be lost along with the rest of the resource.`)

	deleteExample = templates.Examples(`
		# Delete a pod with minimal delay
		kubectl delete pod foo --now

		# Force delete a pod on a dead node
		kubectl delete pod foo --grace-period=0 --force

		# Delete a pod with UID 1234-56-7890-234234-456456.
		kubectl delete pod 1234-56-7890-234234-456456

		# Delete all pods
		kubectl delete pods --all

		# Delete pods and services with same names "baz" and "foo"
		kubectl delete pod,service baz foo

		# Delete pods and services with label name=myLabel.
		kubectl delete pods,services -l name=myLabel
		`)
)

//NewDeleteCmd returns a new command that wraps the kubectl delete command
//it finds the target pod and pass the arguments to the wrapped command
func NewDeleteCmd() *cobra.Command {
	settings := config.C

	handler := &DeletePodCmdHandle{}
	handler.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	handler.writer = os.Stdout

	command := &cobra.Command{
		Use:     "delete ([-f FILENAME] | TYPE [(NAME | -l label | --all)])",
		Short:   "Delete resources by filenames, stdin, resources and names, or by resources and label selector",
		Long:    deleteLong,
		Example: deleteExample,
		Run: func(cmd *cobra.Command, args []string) {
			validateConfig()

			checkErr(handler.Complete(cmd, args, settings))
			checkErr(handler.Validate())
			checkErr(handler.Handle(args))
		},
	}

	//kubectl delete cmd options
	command.Flags().StringVarP(&handler.options.selector, "selector", "l", "", "Selector (label query) to filter on.")
	command.Flags().BoolVar(&handler.options.all, "all", false, "[-all] to select all the specified resources.")
	command.Flags().BoolVar(&handler.options.ignoreNotFound, "ignore-not-found", false, "Treat \"resource not found\" as a successful delete. Defaults to \"true\" when --all is specified.")
	command.Flags().BoolVar(&handler.options.cascade, "cascade", true, "If true, cascade the deletion of the resources managed by this resource (e.g. Pods created by a ReplicationController).  Default true.")
	command.Flags().IntVar(&handler.options.gracePeriod, "grace-period", -1, "Period of time in seconds given to the resource to terminate gracefully. Ignored if negative.")
	command.Flags().BoolVar(&handler.options.now, "now", false, "If true, resources are signaled for immediate shutdown (same as --grace-period=1).")
	command.Flags().BoolVar(&handler.options.force, "force", false, "Immediate deletion of some resources may result in inconsistency or data loss and requires confirmation.")
	command.Flags().DurationVar(&handler.options.timeout, "timeout", 0, "The length of time to wait before giving up on a delete, zero means determine a timeout from the size of the object")

	//cp tool added options
	environment, err := settings.GetString(config.KubeEnvironmentName)
	checkErr(err)
	//cp-remote specific cmd options
	command.Flags().StringVarP(&handler.options.environment, config.KubeEnvironmentName, "e", environment, "The full remote environment name")

	return command
}

//DeletePodCmdHandle holds the information required to delete a pod
type DeletePodCmdHandle struct {
	kubeCtlInit kubectlapi.KubeCtlInitializer
	writer      io.Writer
	options     deletePodCmdOptions
	argsIn      []string
}

type deletePodCmdOptions struct {
	environment, selector                    string
	all, ignoreNotFound, cascade, now, force bool
	gracePeriod                              int
	timeout                                  time.Duration
}

// Complete verifies command line arguments and loads data from the command environment
func (h *DeletePodCmdHandle) Complete(cmd *cobra.Command, argsIn []string, settings *config.Config) error {
	var err error
	if h.options.environment == "" {
		h.options.environment, err = settings.GetString(config.KubeEnvironmentName)
		checkErr(err)
	}
	h.argsIn = argsIn
	return nil
}

// Validate checks that the provided exec options are specified.
func (h *DeletePodCmdHandle) Validate() error {
	if len(strings.Trim(h.options.environment, " ")) == 0 {
		return fmt.Errorf("the environment specified is invalid")
	}
	return nil
}

// Handle set the flags in the kubeclt logs handle command and executes it
func (h *DeletePodCmdHandle) Handle(args []string) error {
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return nil
	}

	clientConfig := kubectlapi.GetNonInteractiveDeferredLoadingClientConfig(user, apiKey, addr, h.options.environment)
	kubeCmdDelete := kubectlcmd.NewCmdDelete(kubectlcmdutil.NewFactory(clientConfig), os.Stdout)

	kubeCmdDelete.Flags().Set("all", strconv.FormatBool(h.options.all))
	kubeCmdDelete.Flags().Set("cascade", strconv.FormatBool(h.options.cascade))
	kubeCmdDelete.Flags().Set("force", strconv.FormatBool(h.options.force))
	kubeCmdDelete.Flags().Set("grace-period", strconv.Itoa(h.options.gracePeriod))
	kubeCmdDelete.Flags().Set("ignore-not-found", strconv.FormatBool(h.options.ignoreNotFound))
	kubeCmdDelete.Flags().Set("now", strconv.FormatBool(h.options.now))
	kubeCmdDelete.Flags().Set("selector", h.options.selector)
	kubeCmdDelete.Flags().Set("timeout", h.options.timeout.String())

	//TODO: Run doesn't return the error, extract the content of NewCmdDelete().Run() and call directly k8s.io/kubernetes/pkg/kubectl/cmd/delete.go::RunDelete()
	//TODO: Log err
	//TODO: Print native kubectl response using CheckErr() in "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	kubeCmdDelete.Run(kubeCmdDelete, h.argsIn)
	return nil
}
