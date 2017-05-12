package cpapi

import (
	"fmt"
	"strconv"

	msgs "github.com/continuouspipe/remote-environment-client/messages"
	"github.com/continuouspipe/remote-environment-client/util"
	"github.com/pkg/errors"
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
	api             DataProvider
	apiEnvironments []APIEnvironment
	qp              util.QuestionPrompter
	err             error
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

//SetAPIKey sets the cp api key
func (q *MultipleChoiceCpEntityQuestioner) Errors() error {
	return q.err
}

//WhichEntities triggers an api request to the cp api based on the arguments and presents to the user a list of options to choose from, it then stores the user response
func (q MultipleChoiceCpEntityQuestioner) WhichEntities() MultipleChoiceCpEntityQuestioner {
	return q.whichProject().whichFlow().whichEnvironment().whichPod()
}

func (q MultipleChoiceCpEntityQuestioner) whichProject() MultipleChoiceCpEntityQuestioner {
	var optionList []MultipleChoiceCpEntityOption
	apiTeams, err := q.api.GetAPITeams()
	if err != nil {
		q.err = errors.Wrap(err, q.err.Error())
	}
	for _, apiTeam := range apiTeams {
		optionList = append(optionList, MultipleChoiceCpEntityOption{
			Name:  apiTeam.Slug,
			Value: apiTeam.Slug,
		})
	}

	if len(optionList) == 0 {
		q.err = errors.Wrap(q.err, msgs.ProjectsNotFound)
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
	apiFlows, err := q.api.GetAPIFlows(q.answers.Project.Value)
	if err != nil {
		q.err = errors.Wrap(err, q.err.Error())
	}
	for _, apiFlow := range apiFlows {
		optionList = append(optionList, MultipleChoiceCpEntityOption{
			Name:  apiFlow.Repository.Name,
			Value: apiFlow.UUID,
		})
	}

	if len(optionList) == 0 {
		q.err = errors.Wrap(q.err, msgs.FlowsNotFound)
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
		q.apiEnvironments, err = q.api.GetAPIEnvironments(q.answers.Flow.Value)
		if err != nil {
			q.err = errors.Wrap(err, q.err.Error())
		}
	}

	for _, apiEnvironment := range q.apiEnvironments {
		optionList = append(optionList, MultipleChoiceCpEntityOption{
			Name:  apiEnvironment.Identifier,
			Value: apiEnvironment.Identifier,
		})
	}

	if len(optionList) == 0 {
		q.err = errors.Wrap(q.err, msgs.EnvironmentsNotFound)
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
		q.apiEnvironments, err = q.api.GetAPIEnvironments(q.answers.Flow.Value)
		if err != nil {
			q.err = errors.Wrap(err, q.err.Error())
		}
	}
	var targetEnv *APIEnvironment
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
		q.err = errors.Wrap(q.err, msgs.RunningPodNotFound)
		return q
	}

	q.answers.Pod = q.ask("pod", optionList)
	return q
}

func (q MultipleChoiceCpEntityQuestioner) ask(entity string, optionList []MultipleChoiceCpEntityOption) MultipleChoiceCpEntityOption {
	if len(optionList) == 0 {
		q.err = errors.Wrap(q.err, "List of option is empty")
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
			q.err = errors.Wrap(err, q.err.Error())
		}
	}
	return optionList[keySelectedAsInt]
}

//Responses returns the struct containing the answer selected by the user
func (q MultipleChoiceCpEntityQuestioner) Responses() MultipleChoiceCpEntityAnswer {
	return q.answers
}
