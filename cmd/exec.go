package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/errors"
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
	"github.com/continuouspipe/remote-environment-client/kubectlapi/pods"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/util"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	kubectlcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

var execExample = fmt.Sprintf(`
# execute -l -all on the web pod
%[1]s ex -- ls -all

# execute -l -all on the web pod overriding the project-key and remote-branch
%[1]s ex -e techup-dev-user -s web -- ls -all
`, config.AppName)

//NewExecCmd return a cobra command struct pointer which on Run, if required it prepares the config so we can reach the pod and
//then uses a command handler to execute the command specified in the arguments
func NewExecCmd() *cobra.Command {
	settings := config.C
	handler := newExecHandle()

	var interactive bool
	var flowID string

	bashcmd := &cobra.Command{
		Use:     "exec",
		Aliases: []string{"ex"},
		Short:   "Execute a command on a container",
		Long: `To execute a command on a container without first getting a bash session use
the exec command. The command and its arguments need to follow --`,
		Example: execExample,
		Run: func(cmd *cobra.Command, args []string) {
			podsFinder := pods.NewKubePodsFind()
			podsFilter := pods.NewKubePodsFilter()
			local := exec.NewLocal()

			defaultEnvironment, err := settings.GetString(config.KubeEnvironmentName)
			checkErr(err)
			defaultService, err := settings.GetString(config.Service)
			checkErr(err)

			if interactive == true {
				cplogs.V(5).Infoln("exec in interactive mode")
				//make sure config has an api key and a cp user set
				initInteractiveH := NewInitInteractiveHandler(false)
				initInteractiveH.SetWriter(ioutil.Discard)
				err := initInteractiveH.Complete(args)
				checkErr(err)
				err = initInteractiveH.Validate()
				checkErr(err)
				err = initInteractiveH.Handle()
				checkErr(err)

				if flowID == "" && handler.environment == "" && handler.service == "" {
					//guide the user to choose the right pod they want to target
					questioner := NewMultipleChoiceCpEntityQuestioner()
					apiKey, _ := settings.GetString(config.ApiKey)
					questioner.api.SetApiKey(apiKey)
					_, flowID, handler.environment, handler.service = questioner.WhichEntities().Responses()
					checkErr(questioner.errors)

					suggestedFlags := color.GreenString("-i -e %s -f %s -s %s", handler.environment, flowID, handler.service)
					fmt.Printf(msgs.InteractiveModeSuggestingFlags, suggestedFlags)
				}

				//alter the configuration so that we connect to the flow and environment specified by the user
				err = newInteractiveModeH().findTargetClusterAndApplyToConfig(flowID, handler.environment)
				checkErr(err)
			} else {
				//apply default values
				if handler.environment == "" {
					handler.environment = defaultEnvironment
				}
				if handler.service == "" {
					handler.service = defaultService
				}
			}

			checkErr(handler.complete(args, settings))
			checkErr(handler.validate())
			checkErr(handler.handle(podsFinder, podsFilter, local))
		},
	}

	bashcmd.PersistentFlags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode requires the flags: --environment --service --flow to be specified")

	bashcmd.PersistentFlags().StringVarP(&handler.environment, config.KubeEnvironmentName, "e", "", "The full remote environment name")
	bashcmd.PersistentFlags().StringVarP(&handler.service, config.Service, "s", "", "The service to use (e.g.: web, mysql)")
	bashcmd.PersistentFlags().StringVarP(&flowID, config.FlowId, "f", "", "The flow to use")

	return bashcmd
}

type execHandle struct {
	args        []string
	config      config.ConfigProvider
	environment string
	service     string
	kubeCtlInit kubectlapi.KubeCtlInitializer
}

func newExecHandle() *execHandle {
	p := &execHandle{}
	p.config = config.C
	p.kubeCtlInit = kubectlapi.NewKubeCtlInit()
	return p
}

// complete verifies command line arguments and loads data from the command environment
func (h *execHandle) complete(argsIn []string, conf config.ConfigProvider) error {
	h.args = argsIn
	h.config = conf
	var err error
	if h.environment == "" {
		h.environment, err = conf.GetString(config.KubeEnvironmentName)
		checkErr(err)
	}
	if h.service == "" {
		h.service, err = conf.GetString(config.Service)
		checkErr(err)
	}
	return nil
}

// validate checks that the provided bash options are specified.
func (h *execHandle) validate() error {
	if len(strings.Trim(h.environment, " ")) == 0 {
		return fmt.Errorf("the environment specified is invalid")
	}
	if len(strings.Trim(h.service, " ")) == 0 {
		return fmt.Errorf("the service specified is invalid")
	}
	return nil
}

// handle opens a bash console against a pod.
func (h *execHandle) handle(podsFinder pods.Finder, podsFilter pods.Filter, executor exec.Executor) error {
	addr, user, apiKey, err := h.kubeCtlInit.GetSettings()
	if err != nil {
		return nil
	}

	podsList, err := podsFinder.FindAll(user, apiKey, addr, h.environment)
	if err != nil {
		return err
	}

	pod := podsFilter.List(*podsList).ByService(h.service).ByStatus("Running").ByStatusReason("Running").First()
	if pod == nil {
		return fmt.Errorf(fmt.Sprintf(msgs.NoActivePodsFoundForSpecifiedServiceName, h.service))
	}

	clientConfig := kubectlapi.GetNonInteractiveDeferredLoadingClientConfig(user, apiKey, addr, h.environment)
	kubeCmdExec := kubectlcmd.NewCmdExec(kubectlcmdutil.NewFactory(clientConfig), os.Stdin, os.Stdout, os.Stderr)
	kubeCmdExec.Flags().Set("pod", pod.GetName())
	kubeCmdExec.Flags().Set("stdin", "true")
	kubeCmdExec.Flags().Set("tty", "true")

	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {

		envTerm := os.Getenv("TERM")
		if envTerm == "" {
			envTerm = "xterm"
		}

		//ensure that the TERM environment variable is set
		//Work-around to be removed when kubernetes and docker fix the issue.
		//See docker/docker#26461 and kubernetes/kubernetes/issues/28280
		h.args = append([]string{"env", "TERM=" + envTerm}, h.args...)
	}

	kubeCmdExec.Run(kubeCmdExec, h.args)
	return nil
}

type interactiveModeHandler interface {
	Handle(flowID string, environment string) error
}

type interactiveModeH struct {
	config config.ConfigProvider
	api    cpapi.CpApiProvider
}

func newInteractiveModeH() *interactiveModeH {
	p := &interactiveModeH{}
	p.config = config.C
	p.api = cpapi.NewCpApi()
	return p
}

func (h interactiveModeH) findTargetClusterAndApplyToConfig(flowID string, targetEnvironment string) error {
	apiKey, err := h.config.GetString(config.ApiKey)
	if err != nil {
		return err
	}

	h.api.SetApiKey(apiKey)

	environments, el := h.api.GetApiEnvironments(flowID)
	if el != nil {
		return el
	}

	clusterIdentifier := ""
	for _, environment := range environments {
		if environment.Identifier == targetEnvironment {
			clusterIdentifier = environment.Cluster
		}
	}

	if clusterIdentifier == "" {
		return fmt.Errorf("The specified environment %s has not been found on the flow %s. Please check that you have inserted the correct flowID and environment at https://ui.continuouspipe.io/project/")
	}

	//set the not persistent config information (do not save the local config in interactive mode)
	h.config.Set(config.CpKubeProxyEnabled, true)
	h.config.Set(config.FlowId, flowID)
	cplogs.V(5).Infof("interactive mode: flow set to %s", flowID)
	h.config.Set(config.ClusterIdentifier, clusterIdentifier)
	cplogs.V(5).Infof("interactive mode: cluster found and is set to %s", clusterIdentifier)
	return nil
}

type MultipleChoiceCpEntityOption struct {
	Value string
	Name  string
}

//MultipleChoiceCpEntityQuestioner presents the user with a list of option for cp entities (project, flows, environments, pods)
//to choose from and requests to select a single one
type MultipleChoiceCpEntityQuestioner struct {
	answers         map[string]MultipleChoiceCpEntityOption
	api             cpapi.CpApiProvider
	errors          errors.ErrorListProvider
	apiEnvironments []cpapi.ApiEnvironment
	qp              util.QuestionPrompter
}

//NewMultipleChoiceCpEntityQuestioner returns a *MultipleChoiceCpEntityQuestioner struct
func NewMultipleChoiceCpEntityQuestioner() *MultipleChoiceCpEntityQuestioner {
	q := &MultipleChoiceCpEntityQuestioner{}
	q.answers = map[string]MultipleChoiceCpEntityOption{
		"project":     {},
		"flowID":      {},
		"environment": {},
		"pod":         {},
	}
	q.api = cpapi.NewCpApi()
	q.errors = errors.NewErrorList()
	q.qp = util.NewQuestionPrompt()
	return q
}

//WhichEntities triggers an api request to the cp api based on the arguments and presents to the user a list of options
//to choose from, it then stores the user response
func (q MultipleChoiceCpEntityQuestioner) WhichEntities() MultipleChoiceCpEntityQuestioner {
	q.WhichProject().whichFlow().whichEnvironment().whichPod()
	return q
}

func (q MultipleChoiceCpEntityQuestioner) WhichProject() MultipleChoiceCpEntityQuestioner {
	var optionList []MultipleChoiceCpEntityOption
	apiTeams, err := q.api.GetApiTeams()
	if err != nil {
		q.errors.Add(err)
	}
	for _, apiTeam := range apiTeams {
		optionList = append(optionList, MultipleChoiceCpEntityOption{
			Name:  apiTeam.Slug,
			Value: apiTeam.Slug,
		})
	}

	if len(optionList) == 0 {
		q.errors.AddErrorf(msgs.ProjectsNotFound)
		return q
	}

	q.answers["project"] = q.ask("project", optionList)
	return q
}

func (q MultipleChoiceCpEntityQuestioner) whichFlow() MultipleChoiceCpEntityQuestioner {
	if q.answers["project"].Value == "" {
		return q
	}

	var optionList []MultipleChoiceCpEntityOption
	apiFlows, err := q.api.GetApiFlows(q.answers["project"].Value)
	if err != nil {
		q.errors.Add(err)
	}
	for _, apiFlow := range apiFlows {
		optionList = append(optionList, MultipleChoiceCpEntityOption{
			Name:  apiFlow.Repository.Name,
			Value: apiFlow.Uuid,
		})
	}

	if len(optionList) == 0 {
		q.errors.AddErrorf(msgs.FlowsNotFound)
		return q
	}

	q.answers["flowID"] = q.ask("flowID", optionList)
	return q
}

func (q MultipleChoiceCpEntityQuestioner) whichEnvironment() MultipleChoiceCpEntityQuestioner {
	if q.answers["flowID"].Value == "" {
		return q
	}

	var optionList []MultipleChoiceCpEntityOption
	if len(q.apiEnvironments) == 0 {
		var err error
		q.apiEnvironments, err = q.api.GetApiEnvironments(q.answers["flowID"].Value)
		if err != nil {
			q.errors.Add(err)
		}
	}

	for _, apiEnvironment := range q.apiEnvironments {
		optionList = append(optionList, MultipleChoiceCpEntityOption{
			Name:  apiEnvironment.Identifier,
			Value: apiEnvironment.Identifier,
		})
	}

	if len(optionList) == 0 {
		q.errors.AddErrorf(msgs.EnvironmentsNotFound)
		return q
	}

	q.answers["environment"] = q.ask("environment", optionList)
	return q
}

func (q MultipleChoiceCpEntityQuestioner) whichPod() MultipleChoiceCpEntityQuestioner {
	if q.answers["flowID"].Value == "" {
		return q
	}

	var optionList []MultipleChoiceCpEntityOption
	if len(q.apiEnvironments) == 0 {
		var err error
		q.apiEnvironments, err = q.api.GetApiEnvironments(q.answers["flowID"].Value)
		if err != nil {
			q.errors.Add(err)
		}
	}
	var targetEnv *cpapi.ApiEnvironment
	for _, apiEnvironment := range q.apiEnvironments {
		if apiEnvironment.Identifier == q.answers["environment"].Value {
			targetEnv = &apiEnvironment
			break
		}
	}
	if targetEnv != nil {
		for _, pod := range targetEnv.Components {
			optionList = append(optionList, MultipleChoiceCpEntityOption{
				Name:  pod.Name,
				Value: pod.Name,
			})
		}
	}

	if len(optionList) == 0 {
		fmt.Println(msgs.RunningPodNotFound)
		return q
	}

	q.answers["pod"] = q.ask("pod", optionList)
	return q
}

func (q MultipleChoiceCpEntityQuestioner) ask(entity string, optionList []MultipleChoiceCpEntityOption) MultipleChoiceCpEntityOption {
	if len(optionList) == 0 {
		q.errors.AddErrorf("List of option is empty")
		return MultipleChoiceCpEntityOption{}
	}
	if len(optionList) == 1 {
		for _, option := range optionList {
			return option
		}
	}

	printedOptions := ""
	for key, v := range optionList {
		printedOptions = printedOptions + fmt.Sprintf("[%d] %s\n", key, v.Name)
	}

	question := fmt.Sprintf("Which %s would you like to use?\n"+
		"%s\n"+
		"Select an option from 0 to %d: ", entity, printedOptions, len(optionList)-1)

	keySelected := q.qp.RepeatUntilValid(question, func(answer string) (bool, error) {
		for key, _ := range optionList {
			if strconv.Itoa(key) == answer {
				return true, nil
			}
		}
		return false, fmt.Errorf("Please select an option between [0-%d]", len(optionList))

	})
	keySelectedAsInt, err := strconv.Atoi(keySelected)
	if err != nil {
		q.errors.Add(err)
	}
	return optionList[keySelectedAsInt]
}

//ApplyDefault sets an aswer as if it was given by the user. Can be used to fetch the environment without asking to the user to specify the the Project and FlowId
func (q MultipleChoiceCpEntityQuestioner) ApplyDefault(arg string, option MultipleChoiceCpEntityOption) MultipleChoiceCpEntityQuestioner {
	q.answers[arg] = option
	return q
}

//Responses returns the answers
func (q MultipleChoiceCpEntityQuestioner) Responses() (project string, flowID string, environment string, pod string) {
	return q.answers["project"].Value, q.answers["flowID"].Value, q.answers["environment"].Value, q.answers["pod"].Value
}
