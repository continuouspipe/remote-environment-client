package cpapi

import (
	"fmt"
	"net/http"
	"strconv"

	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/session"
	"github.com/continuouspipe/remote-environment-client/util"
	"github.com/pkg/errors"
)

const errorFailedToGetFlowList = "failed to get the flow list"
const errorFailedToGetTeamList = "failed to get the teams list"
const questionerMultiSelectError = "error in the questioner multi select logic"

//MultipleChoiceCpEntityOption holds an option that will be presented to the user
type MultipleChoiceCpEntityOption struct {
	Value  string
	Name   string
	Object interface{}
}

//MultipleChoiceCpEntityAnswer for each entity holds the option selected by the user
type MultipleChoiceCpEntityAnswer struct {
	Project     MultipleChoiceCpEntityOption
	Flow        MultipleChoiceCpEntityOption
	Environment MultipleChoiceCpEntityOption
	Pod         MultipleChoiceCpEntityOption
}

//MultipleChoiceCpEntityQuestioner presents the user with a list of option for cp entities (project, flows, environments, pods)
//to choose from and requests to select a single one
type MultipleChoiceCpEntityQuestioner struct {
	api DataProvider
	qp  util.QuestionPrompter
}

//NewMultipleChoiceCpEntityQuestioner returns a *MultipleChoiceCpEntityQuestioner struct
func NewMultipleChoiceCpEntityQuestioner() *MultipleChoiceCpEntityQuestioner {
	q := &MultipleChoiceCpEntityQuestioner{}
	q.api = NewCpAPI()
	q.qp = util.NewQuestionPrompt()
	return q
}

//SetAPIKey sets the cp api key
func (q *MultipleChoiceCpEntityQuestioner) SetAPIKey(apiKey string) {
	q.api.SetAPIKey(apiKey)
}

//WhichEntities triggers an api request to the cp api based on the arguments and presents to the user a list of options to choose from, it then stores the user response
func (q *MultipleChoiceCpEntityQuestioner) WhichEntities() (project APITeam, flow APIFlow, environment APIEnvironment, pod APIComponent, suggestion string, err error) {
	project, suggestion, err = q.whichProject()
	if err != nil {
		return
	}
	flow, suggestion, err = q.whichFlow(project.Slug)
	if err != nil {
		return
	}
	environment, suggestion, err = q.whichEnvironment(flow.UUID)
	if err != nil {
		return
	}
	pod, suggestion, err = q.whichPod(environment)
	if err != nil {
		return
	}
	return
}

func (q MultipleChoiceCpEntityQuestioner) whichProject() (apiTeam APITeam, suggestion string, err error) {
	var optionList []MultipleChoiceCpEntityOption
	apiTeams, err := q.api.GetAPITeams()
	if err != nil {
		return APITeam{}, fmt.Sprintf(msgs.SuggestionGetAPITeamsFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToGetTeamList).String())
	}
	for _, apiTeam := range apiTeams {
		optionList = append(optionList, MultipleChoiceCpEntityOption{
			Name:   apiTeam.Slug,
			Value:  apiTeam.Slug,
			Object: apiTeam,
		})
	}

	if len(optionList) == 0 {
		return APITeam{}, fmt.Sprintf(msgs.SuggestionProjectsListEmpty, session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.ProjectsNotFound).String())
	}

	opt, err := q.ask("project", optionList)
	if err != nil {
		return APITeam{}, fmt.Sprintf(msgs.SuggestionQuestionerMultiSelectError, session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, questionerMultiSelectError).String())
	}
	return opt.Object.(APITeam), "", nil
}

func (q MultipleChoiceCpEntityQuestioner) whichFlow(projectID string) (apiFlow APIFlow, suggestion string, err error) {
	var optionList []MultipleChoiceCpEntityOption
	apiFlows, err := q.api.GetAPIFlows(projectID)
	if err != nil {
		return APIFlow{}, fmt.Sprintf(msgs.SuggestionGetFlowListFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToGetFlowList).String())
	}
	for _, apiFlow := range apiFlows {
		optionList = append(optionList, MultipleChoiceCpEntityOption{
			Name:   apiFlow.Repository.Name,
			Value:  apiFlow.UUID,
			Object: apiFlow,
		})
	}

	if len(optionList) == 0 {
		return APIFlow{}, fmt.Sprintf(msgs.SuggestionFlowListEmpty, projectID, session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.FlowsNotFound).String())
	}

	opt, err := q.ask("flow", optionList)
	if err != nil {
		return APIFlow{}, fmt.Sprintf(msgs.SuggestionQuestionerMultiSelectError, session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, questionerMultiSelectError).String())
	}
	return opt.Object.(APIFlow), "", nil
}

func (q MultipleChoiceCpEntityQuestioner) whichEnvironment(flowID string) (environment APIEnvironment, suggestion string, err error) {
	var optionList []MultipleChoiceCpEntityOption

	apiEnvironments, err := q.api.GetAPIEnvironments(flowID)
	if err != nil {
		return APIEnvironment{}, fmt.Sprintf(msgs.SuggestionGetApiEnvironmentsFailed, session.CurrentSession.SessionID), errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, ErrorFailedToGetEnvironmentsList).String())
	}

	for _, apiEnvironment := range apiEnvironments {
		optionList = append(optionList, MultipleChoiceCpEntityOption{
			Name:   apiEnvironment.Identifier,
			Value:  apiEnvironment.Identifier,
			Object: apiEnvironment,
		})
	}

	if len(optionList) == 0 {
		return APIEnvironment{}, fmt.Sprintf(msgs.SuggestionEnvironmentListEmpty, flowID, session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.EnvironmentsNotFound).String())
	}

	opt, err := q.ask("environment", optionList)
	if err != nil {
		return APIEnvironment{}, fmt.Sprintf(msgs.SuggestionQuestionerMultiSelectError, session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, questionerMultiSelectError).String())
	}

	return opt.Object.(APIEnvironment), "", nil
}

func (q *MultipleChoiceCpEntityQuestioner) whichPod(environment APIEnvironment) (pod APIComponent, suggestion string, err error) {
	var optionList []MultipleChoiceCpEntityOption

	for _, pod := range environment.Components {
		optionList = append(optionList, MultipleChoiceCpEntityOption{
			Name:   pod.Name,
			Value:  pod.Name,
			Object: pod,
		})
	}

	if len(optionList) == 0 {
		return APIComponent{}, fmt.Sprintf(msgs.SuggestionRunningPodNotFound, environment.Identifier, session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msgs.RunningPodNotFound).String())
	}

	opt, err := q.ask("pod", optionList)
	if err != nil {
		return APIComponent{}, fmt.Sprintf(msgs.SuggestionQuestionerMultiSelectError, session.CurrentSession.SessionID), errors.New(cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, questionerMultiSelectError).String())
	}
	return opt.Object.(APIComponent), "", nil
}

func (q MultipleChoiceCpEntityQuestioner) ask(entity string, optionList []MultipleChoiceCpEntityOption) (option MultipleChoiceCpEntityOption, err error) {
	if len(optionList) == 0 {
		return MultipleChoiceCpEntityOption{}, errors.Wrap(err, "List of option is empty")
	}
	if len(optionList) == 1 {
		return optionList[0], nil
	}

	printedOptions := ""
	for key, v := range optionList {
		printedOptions = printedOptions + fmt.Sprintf("[%d] %s\n", key, v.Name)
	}

	question := fmt.Sprintf("Which %s would you like to use?\n"+
		"%s\n"+
		"Please insert the option in full or a type its corrispondent value between [0-%d]:", entity, printedOptions, len(optionList)-1)

	optionSelected := q.qp.RepeatUntilValid(question, func(answer string) (bool, error) {
		for key, option := range optionList {
			if option.Name == answer {
				return true, nil
			}
			if strconv.Itoa(key) == answer {
				return true, nil
			}
		}
		return false, fmt.Errorf("Please insert the option in full or a type its correspondent value between [0-%d]", len(optionList))

	})

	keySelectedAsInt, err := strconv.Atoi(optionSelected)
	if err != nil {
		keySelectedAsInt = -1
		for key, v := range optionList {
			if v.Name == optionSelected {
				keySelectedAsInt = key
				break
			}
		}
		if keySelectedAsInt == -1 {
			return MultipleChoiceCpEntityOption{}, errors.New("invalid option key selected")
		}
	}
	return optionList[keySelectedAsInt], nil
}
