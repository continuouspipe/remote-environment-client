package cpapi

import (
	"fmt"
	"strconv"

	"github.com/continuouspipe/remote-environment-client/errors"
	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/util"
)

//MultipleChoiceCpEntityOption holds an option that will be presented to the user
type MultipleChoiceCpEntityOption struct {
	Value string
	Name  string
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
	answers         MultipleChoiceCpEntityAnswer
	api             CpApiProvider
	errors          errors.ErrorListProvider
	apiEnvironments []ApiEnvironment
	qp              util.QuestionPrompter
}

//NewMultipleChoiceCpEntityQuestioner returns a *MultipleChoiceCpEntityQuestioner struct
func NewMultipleChoiceCpEntityQuestioner() *MultipleChoiceCpEntityQuestioner {
	q := &MultipleChoiceCpEntityQuestioner{}
	q.api = NewCpApi()
	q.errors = errors.NewErrorList()
	q.qp = util.NewQuestionPrompt()
	return q
}

//SetApiKey sets the cp api key
func (q *MultipleChoiceCpEntityQuestioner) SetApiKey(apiKey string) {
	q.api.SetApiKey(apiKey)
}

//SetApiKey sets the cp api key
func (q *MultipleChoiceCpEntityQuestioner) Errors() errors.ErrorListProvider {
	return q.errors
}

//WhichEntities triggers an api request to the cp api based on the arguments and presents to the user a list of options to choose from, it then stores the user response
func (q MultipleChoiceCpEntityQuestioner) WhichEntities() MultipleChoiceCpEntityQuestioner {
	return q.whichProject().whichFlow().whichEnvironment().whichPod()
}

func (q MultipleChoiceCpEntityQuestioner) whichProject() MultipleChoiceCpEntityQuestioner {
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

	q.answers.Project = q.ask("project", optionList)
	return q
}

func (q MultipleChoiceCpEntityQuestioner) whichFlow() MultipleChoiceCpEntityQuestioner {
	if q.answers.Project.Value == "" {
		return q
	}

	var optionList []MultipleChoiceCpEntityOption
	apiFlows, err := q.api.GetApiFlows(q.answers.Project.Value)
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

	q.answers.Flow = q.ask("flow", optionList)
	return q
}

func (q MultipleChoiceCpEntityQuestioner) whichEnvironment() MultipleChoiceCpEntityQuestioner {
	if q.answers.Flow.Value == "" {
		return q
	}

	var optionList []MultipleChoiceCpEntityOption
	if len(q.apiEnvironments) == 0 {
		var err error
		q.apiEnvironments, err = q.api.GetApiEnvironments(q.answers.Flow.Value)
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

	q.answers.Environment = q.ask("environment", optionList)
	return q
}

func (q MultipleChoiceCpEntityQuestioner) whichPod() MultipleChoiceCpEntityQuestioner {
	if q.answers.Flow.Value == "" {
		return q
	}

	var optionList []MultipleChoiceCpEntityOption
	if len(q.apiEnvironments) == 0 {
		var err error
		q.apiEnvironments, err = q.api.GetApiEnvironments(q.answers.Flow.Value)
		if err != nil {
			q.errors.Add(err)
		}
	}
	var targetEnv *ApiEnvironment
	for _, apiEnvironment := range q.apiEnvironments {
		if apiEnvironment.Identifier == q.answers.Environment.Value {
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
		q.errors.AddErrorf(msgs.RunningPodNotFound)
		return q
	}

	q.answers.Pod = q.ask("pod", optionList)
	return q
}

func (q MultipleChoiceCpEntityQuestioner) ask(entity string, optionList []MultipleChoiceCpEntityOption) MultipleChoiceCpEntityOption {
	if len(optionList) == 0 {
		q.errors.AddErrorf("List of option is empty")
		return MultipleChoiceCpEntityOption{}
	}
	if len(optionList) == 1 {
		return optionList[0]
	}

	printedOptions := ""
	for key, v := range optionList {
		printedOptions = printedOptions + fmt.Sprintf("[%d] %s\n", key, v.Name)
	}

	question := fmt.Sprintf("Which %s would you like to use?\n"+
		"%s\n"+
		"Select an option from 0 to %d: ", entity, printedOptions, len(optionList)-1)

	keySelected := q.qp.RepeatUntilValid(question, func(answer string) (bool, error) {
		for key := range optionList {
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

//Responses returns the struct containing the answer selected by the user
func (q MultipleChoiceCpEntityQuestioner) Responses() MultipleChoiceCpEntityAnswer {
	return q.answers
}
