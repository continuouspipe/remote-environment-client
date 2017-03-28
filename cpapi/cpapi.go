package cpapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"bytes"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/errors"
	"io"
)

type CpApiProvider interface {
	SetApiKey(apiKey string)
	GetApiTeams() ([]ApiTeam, error)
	GetApiUser(user string) (*ApiUser, error)
	GetApiEnvironments(flowId string) ([]ApiEnvironment, errors.ErrorListProvider)
	GetRemoteEnvironmentStatus(flowId string, environmentId string) (*ApiRemoteEnvironmentStatus, errors.ErrorListProvider)
	RemoteEnvironmentBuild(remoteEnvironmentFlowID string, gitBranch string) error
	CancelRunningTide(flowId string, remoteEnvironmentId string) error
	RemoteEnvironmentRunningAndExists(flowId string, environmentId string) (bool, errors.ErrorListProvider)
	RemoteEnvironmentDestroy(flowId string, environment string, cluster string) error
	RemoteDevelopmentEnvironmentDestroy(flowId string, remoteEnvironmentId string) error
	CancelTide(tideId string) error
}

type CpApi struct {
	client *http.Client
	apiKey string
}

func NewCpApi() *CpApi {
	clusterInfo := &CpApi{}
	clusterInfo.client = &http.Client{}
	return clusterInfo
}

const RemoteEnvironmentRunning = "Running"
const RemoteEnvironmentTideFailed = "TideFailed"
const RemoteEnvironmentTideRunning = "TideRunning"
const RemoteEnvironmentTideNotStarted = "NotStarted"

const TideRunning = "running"

type ApiTeam struct {
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	BucketUuid string `json:"bucket_uuid"`

	//Should be []ApiMembership although there is a bug on the api where a list of object with keys "1", "2" is returned
	//instead of being a json array
	Memberships []interface{} `json:"memberships"`
}

type ApiMembership struct {
	Team        ApiTeam  `json:"team"`
	User        ApiUser  `json:"user"`
	Permissions []string `json:"permissions"`
}

type ApiUser struct {
	Username   string   `json:"username"`
	Email      string   `json:"email"`
	BucketUuid string   `json:"bucket_uuid"`
	Roles      []string `json:"roles"`
}

type ApiCluster struct {
	Identifier string `json:"identifier"`
	Address    string `json:"address"`
	Version    string `json:"version"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Type       string `json:"type"`
}

type ApiRemoteEnvironmentStatus struct {
	Status              string              `json:"status"`
	KubeEnvironmentName string              `json:"environment_name"`
	ClusterIdentifier   string              `json:"cluster_identifier"`
	PublicEndpoints     []ApiPublicEndpoint `json:"public_endpoints"`
	LastTide            ApiTide             `json:"last_tide"`
}

type ApiPublicEndpoint struct {
	Address string                  `json:"address"`
	Name    string                  `json:"name"`
	Ports   []ApiPublicEndpointPort `json:"ports"`
}

type ApiPublicEndpointPort struct {
	Number   int    `json:"number"`
	Protocol string `json:"protocol"`
}

type ApiTide struct {
	CodeReference  ApiCodeReference `json:"code_reference"`
	Configuration  interface{}      `json:"configuration"`
	CreationDate   string           `json:"creation_date"`
	FinishDate     string           `json:"finish_date"`
	FlowUuid       string           `json:"flow_uuid"`
	GenerationUuid string           `json:"generation_uuid"`
	LogId          string           `json:"log_id"`
	StartDate      string           `json:"start_date"`
	Status         string           `json:"status"`
	Tasks          []interface{}    `json:"tasks"`
	Team           ApiTideTeam      `json:"team"`
	User           interface{}      `json:"user"`
	Uuid           string           `json:"uuid"`
}

type ApiTideTeam struct {
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	BucketUuid string `json:"bucket_uuid"`
}

type ApiCodeReference struct {
	Branch         string      `json:"branch"`
	CodeRepository interface{} `json:"code_repository"`
	Sha1           string      `json:"sha1"`
}

type ApiEnvironment struct {
	Cluster    string        `json:"cluster"`
	Components []interface{} `json:"components"`
	Identifier string        `json:"identifier"`
}

func (c *CpApi) SetApiKey(apiKey string) {
	c.apiKey = apiKey
}

func (c CpApi) GetApiTeams() ([]ApiTeam, error) {
	el := errors.NewErrorList()
	el.AddErrorf("error when getting the list of the teams using the CP Api")

	if c.apiKey == "" {
		el.AddErrorf("Api key not provided.")
		return nil, el
	}

	u, err := c.getAuthenticatorURL()
	if err != nil {
		el.Add(err)
		return nil, el
	}
	u.Path = "/api/teams"

	cplogs.V(5).Infof("getting api teams info on cp using url %s", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		el.Add(err)
		return nil, el
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	respBody, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof("failed to get the response body for request %s", u.String())
		cplogs.Flush()
		el.Add(elr.Items()...)
		return nil, el
	}

	teams := make([]ApiTeam, 0)
	err = json.Unmarshal(respBody, teams)
	if err != nil {
		cplogs.V(4).Infof("error running Unmarshal() on response body %s", respBody)
		el.Add(err)
		return nil, el
	}

	return teams, nil
}

func (c CpApi) GetApiUser(user string) (*ApiUser, error) {
	el := errors.NewErrorList()
	el.AddErrorf("error when getting the CP User using the CP Api")

	if c.apiKey == "" {
		el.AddErrorf("Api key not provided.")
		return nil, el
	}

	u, err := c.getAuthenticatorURL()
	if err != nil {
		el.Add(err)
		return nil, el
	}
	u.Path = "/api/user/" + user

	cplogs.V(5).Infof("getting user info on cp using url %s", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		el.Add(err)
		return nil, el
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	respBody, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof("failed to get the response body for request %s", u.String())
		cplogs.Flush()
		el.Add(elr.Items()...)
		return nil, el
	}

	apiUserResponse := &ApiUser{}
	err = json.Unmarshal(respBody, apiUserResponse)
	if err != nil {
		cplogs.V(4).Infof("error running Unmarshal() on response body %s", respBody)
		el.Add(err)
		return nil, el
	}

	return apiUserResponse, nil
}

func (c CpApi) GetApiEnvironments(flowId string) ([]ApiEnvironment, errors.ErrorListProvider) {
	el := errors.NewErrorList()
	el.AddErrorf("error when getting the list of environments using the CP Api")

	if c.apiKey == "" {
		el.AddErrorf("api key not provided")
		return nil, el
	}

	url, err := c.getRiverURL()
	if err != nil {
		el.AddErrorf("error when getting the river url")
		el.Add(err)
		return nil, el
	}
	url.Path = fmt.Sprintf("/flows/%s/environments", flowId)

	cplogs.V(5).Infof("getting flow environments using url %s", url.String())

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		el.AddErrorf("error when getting a new http request object for fetching the list of environments")
		el.Add(err)
		return nil, el
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	respBody, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		el.Add(elr.Items()...)
		return nil, el
	}

	environments := make([]ApiEnvironment, 0)
	err = json.Unmarshal(respBody, &environments)
	if err != nil {
		errText := fmt.Sprintf("error running Unmarshal() on response body %s", respBody)
		cplogs.V(4).Infof(errText)
		el.AddErrorf(errText)
		el.Add(err)
		return nil, el
	}

	return environments, nil
}

//calls CP Api to retrieve information about the remote environment
func (c CpApi) GetRemoteEnvironmentStatus(flowId string, environmentId string) (*ApiRemoteEnvironmentStatus, errors.ErrorListProvider) {
	el := errors.NewErrorList()
	el.AddErrorf("error when getting the remote environment status using the CP Api")

	if c.apiKey == "" {
		el.AddErrorf("api key not provided")
		return nil, el
	}

	u, err := c.getRiverURL()
	if err != nil {
		el.AddErrorf("error when getting the river url")
		el.Add(err)
		return nil, el
	}
	u.Path = fmt.Sprintf("/flows/%s/development-environments/%s/status", flowId, environmentId)

	cplogs.V(5).Infof("getting remote environment status using url %s", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		el.AddErrorf("error when getting a new http request object for fetching the environment status")
		el.Add(err)
		return nil, el
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	respBody, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof("failed to get the response body for request %s", u.String())
		cplogs.Flush()
		el.Add(elr.Items()...)
		return nil, el
	}

	apiRemoteEnvironment := &ApiRemoteEnvironmentStatus{}
	err = json.Unmarshal(respBody, apiRemoteEnvironment)
	if err != nil {
		errText := fmt.Sprintf("error running Unmarshal() on response body %s", respBody)
		cplogs.V(4).Infof(errText)
		el.AddErrorf(errText)
		el.Add(err)
		return nil, el
	}

	return apiRemoteEnvironment, nil
}

//calls CP API to request to build a new remote environment
func (c CpApi) RemoteEnvironmentBuild(remoteEnvironmentFlowID string, gitBranch string) error {
	el := errors.NewErrorList()
	el.AddErrorf("error when triggering the build for the remote environment using the CP Api")

	if c.apiKey == "" {
		el.AddErrorf("Api key not provided.")
		return el
	}

	u, err := c.getRiverURL()
	if err != nil {
		el.Add(err)
		return el
	}
	u.Path = fmt.Sprintf("/flows/%s/tides", remoteEnvironmentFlowID)

	type requestBody struct {
		BranchName string `json:"branch"`
	}
	reqBodyJson, err := json.Marshal(&requestBody{gitBranch})

	cplogs.V(5).Infof("triggering remote environment build using url %s and payload %s", u.Path, reqBodyJson)

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(reqBodyJson))
	if err != nil {
		el.Add(err)
		return el
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)

	_, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof("failed to get the response body for request %s", u.String())
		cplogs.Flush()
		el.Add(elr.Items()...)
		return el
	}

	return nil
}

//find the tide id associated with the branch and cancel the tide
func (c CpApi) CancelRunningTide(flowId string, remoteEnvironmentId string) error {
	el := errors.NewErrorList()
	el.AddErrorf("error when cancelling the running tide using the CP Api")

	remoteEnv, elr := c.GetRemoteEnvironmentStatus(flowId, remoteEnvironmentId)
	if elr != nil {
		el.Add(elr.Items()...)
		return el
	}

	if remoteEnv.LastTide.Status != TideRunning {
		cplogs.V(5).Infof("TideId %s not running, skipping", remoteEnv.LastTide.Uuid)
		cplogs.Flush()
		return nil
	}

	return c.CancelTide(remoteEnv.LastTide.Uuid)
}

func (c CpApi) CancelTide(tideId string) error {
	el := errors.NewErrorList()
	el.AddErrorf("error when cancelling the tide using the CP Api")

	if c.apiKey == "" {
		el.AddErrorf("Api key not provided.")
		return el
	}

	u, err := c.getRiverURL()
	if err != nil {
		el.Add(err)
		return el
	}
	u.Path = fmt.Sprintf("/tides/%s/cancel", tideId)

	cplogs.V(5).Infof("cancelling tide using url %s", u.String())

	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		el.Add(err)
		return el
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)

	_, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof("failed to get the response body for request %s", u.String())
		cplogs.Flush()
		el.Add(elr.Items()...)
		return el
	}

	return nil
}

//calls CP API to request to destroy the remote environment
func (c CpApi) RemoteEnvironmentDestroy(flowId string, environment string, cluster string) error {
	el := errors.NewErrorList()
	el.AddErrorf("error when requesting the destruction of the environment using the CP Api")

	if c.apiKey == "" {
		el.AddErrorf("Api key not provided.")
		return el
	}

	u, err := c.getRiverURL()
	if err != nil {
		el.Add(err)
		return el
	}
	u.Path = fmt.Sprintf("/flows/%s/environments/%s", flowId, environment)
	u.RawQuery = fmt.Sprintf("cluster=%s", cluster)

	cplogs.V(5).Infof("destroying remote environment using url %s", u.String())

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		el.Add(err)
		return el
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)

	_, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof("failed to get the response body for request %s", u.String())
		cplogs.Flush()
		el.Add(elr.Items()...)
		return el
	}

	return nil
}

func (c CpApi) RemoteEnvironmentRunningAndExists(flowId string, environmentId string) (bool, errors.ErrorListProvider) {
	el := errors.NewErrorList()

	remoteEnv, elr := c.GetRemoteEnvironmentStatus(flowId, environmentId)
	if elr != nil {
		el.Add(elr.Items()...)
		return false, el
	}

	environments, err := c.GetApiEnvironments(flowId)
	if err != nil {
		el.Add(elr.Items()...)
		return false, el
	}
	for _, environment := range environments {
		if environment.Identifier == remoteEnv.KubeEnvironmentName {
			return true, nil
		}
	}
	return true, nil
}

func (c CpApi) RemoteDevelopmentEnvironmentDestroy(flowId string, remoteEnvironmentId string) error {
	el := errors.NewErrorList()
	el.AddErrorf("error when requesting the destruction of the remote environment using the CP Api")

	if c.apiKey == "" {
		el.AddErrorf("Api key not provided.")
		return el
	}

	u, err := c.getRiverURL()
	if err != nil {
		el.Add(err)
		return el
	}
	u.Path = fmt.Sprintf("/flows/%s/development-environments/%s", flowId, remoteEnvironmentId)

	cplogs.V(5).Infof("destroying remote development environment using url %s", u.String())

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		el.Add(err)
		return el
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)

	_, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof("failed to get the response body for request %s", u.String())
		cplogs.Flush()
		el.Add(elr.Items()...)
		return el
	}

	return nil
}

func (c CpApi) getResponseBody(client *http.Client, req *http.Request) ([]byte, errors.ErrorListProvider) {
	el := errors.NewErrorList()
	res, err := client.Do(req)
	if err != nil {
		el.Add(err)
		return nil, el
	}
	defer res.Body.Close()
	if res.StatusCode < 200 && res.StatusCode > 202 {
		el.AddErrorf("error getting response body, status: %d, url: %s", res.StatusCode, req.URL.String())
		return nil, el
	}
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		el.Add(err)
		return nil, el
	}
	return resBody, nil
}

func (c CpApi) getAuthenticatorURL() (*url.URL, error) {
	cpApiAddr, err := config.C.GetString(config.CpAuthenticatorApiAddr)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(cpApiAddr)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (c CpApi) getRiverURL() (*url.URL, error) {
	cpApiAddr, err := config.C.GetString(config.CpRiverApiAddr)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(cpApiAddr)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func PrintPublicEndpoints(writer io.Writer, endpoints []ApiPublicEndpoint) {
	for _, publicEndpoint := range endpoints {
		for _, port := range publicEndpoint.Ports {
			switch port.Number {
			case 80:
				fmt.Fprintf(writer, "%s \t http://%s\n", publicEndpoint.Name, publicEndpoint.Address)
			case 443:
				fmt.Fprintf(writer, "%s \t https://%s\n", publicEndpoint.Name, publicEndpoint.Address)
			default:
				fmt.Fprintf(writer, "%s \t %s:%d\n", publicEndpoint.Name, publicEndpoint.Address, port.Number)
			}
		}
	}
}
