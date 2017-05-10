package cpapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"bytes"
	"io"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	"github.com/pkg/errors"
)

const errorAPIKeyNotProvided = "api key not provided"
const errorFailedToRetrievedAuthenticatorURL = "failed to retrieve the authenticator url"
const errorFailedToRetrievedRiverURL = "failed to retrieve the river url"
const errorFailedToCreateGetRequest = "failed to create a get request"
const errorFailedToCreatePostRequest = "failed to create a post request"
const errorFailedToCreateDeleteRequest = "failed to create a delete request"
const errorResponseStatusCodeUnsuccessful = "failed to get response body, status: %d, url: %s"
const errorParsingJSONResponse = "failed to unparse json response body %s"
const errorFailedToGetResponseBody = "failed to get the response body"
const errorFailedToGetRemoteEnvironmentStatus = "failed to get remote environment status"
const errorFailedToGetEnvironmentsList = "failed to get the environments list"

//DataProvider collects all cp api methods
type DataProvider interface {
	SetAPIKey(apiKey string)
	GetAPITeams() ([]APITeam, error)
	GetAPIFlows(project string) ([]APIFlow, error)
	GetAPIUser(user string) (*APIUser, error)
	GetAPIEnvironments(flowID string) ([]APIEnvironment, error)
	GetRemoteEnvironmentStatus(flowID string, environmentID string) (*APIRemoteEnvironmentStatus, error)
	RemoteEnvironmentBuild(remoteEnvironmentFlowID string, gitBranch string) error
	CancelRunningTide(flowID string, remoteEnvironmentID string) error
	RemoteEnvironmentRunningAndExists(flowID string, environmentID string) (bool, error)
	RemoteEnvironmentDestroy(flowID string, environment string, cluster string) error
	RemoteDevelopmentEnvironmentDestroy(flowID string, remoteEnvironmentID string) error
	CancelTide(tideID string) error
}

//CpAPI holds the dependencies required to do the api calls
type CpAPI struct {
	client *http.Client
	apiKey string
}

//NewCpAPI ctor for the CpAPI
func NewCpAPI() *CpAPI {
	clusterInfo := &CpAPI{}
	clusterInfo.client = &http.Client{}
	return clusterInfo
}

//RemoteEnvironmentRunning is the status for a running remove environment
const RemoteEnvironmentRunning = "Running"

//RemoteEnvironmentTideFailed is the status for a tide that failed
const RemoteEnvironmentTideFailed = "TideFailed"

//RemoteEnvironmentTideRunning is the status for a tide that is running
const RemoteEnvironmentTideRunning = "TideRunning"

//RemoteEnvironmentTideNotStarted is the status for at tide that has not been started yet
const RemoteEnvironmentTideNotStarted = "NotStarted"

//TideRunning is the status of a running tide
const TideRunning = "running"

//APITeam holds the data expected from the cp api for this entity
type APITeam struct {
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	BucketUUID string `json:"bucket_uuid"`

	//Should be []APIMembership although there is a bug on the api where a list of object with keys "1", "2" is returned
	//instead of being a json array
	Memberships []interface{} `json:"memberships"`
}

//APIMembership holds the data expected from the cp api for this entity
type APIMembership struct {
	Team        APITeam  `json:"team"`
	User        APIUser  `json:"user"`
	Permissions []string `json:"permissions"`
}

//APIUser holds the data expected from the cp api for this entity
type APIUser struct {
	Username   string   `json:"username"`
	Email      string   `json:"email"`
	BucketUUID string   `json:"bucket_uuid"`
	Roles      []string `json:"roles"`
}

//APICluster holds the data expected from the cp api for this entity
type APICluster struct {
	Identifier string `json:"identifier"`
	Address    string `json:"address"`
	Version    string `json:"version"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Type       string `json:"type"`
}

//APIRemoteEnvironmentStatus holds the data expected from the cp api for this entity
type APIRemoteEnvironmentStatus struct {
	Status              string              `json:"status"`
	KubeEnvironmentName string              `json:"environment_name"`
	ClusterIdentifier   string              `json:"cluster_identifier"`
	PublicEndpoints     []APIPublicEndpoint `json:"public_endpoints"`
	LastTide            APITide             `json:"last_tide"`
}

//APIPublicEndpoint holds the data expected from the cp api for this entity
type APIPublicEndpoint struct {
	Address string                  `json:"address"`
	Name    string                  `json:"name"`
	Ports   []APIPublicEndpointPort `json:"ports"`
}

//APIPublicEndpointPort holds the data expected from the cp api for this entity
type APIPublicEndpointPort struct {
	Number   int    `json:"number"`
	Protocol string `json:"protocol"`
}

//APITide holds the data expected from the cp api for this entity
type APITide struct {
	CodeReference  APICodeReference `json:"code_reference"`
	Configuration  interface{}      `json:"configuration"`
	CreationDate   string           `json:"creation_date"`
	FinishDate     string           `json:"finish_date"`
	FlowUUID       string           `json:"flow_uuid"`
	GenerationUUID string           `json:"generation_uuid"`
	LogID          string           `json:"log_id"`
	StartDate      string           `json:"start_date"`
	Status         string           `json:"status"`
	Tasks          []interface{}    `json:"tasks"`
	Team           APITideTeam      `json:"team"`
	User           interface{}      `json:"user"`
	UUID           string           `json:"uuid"`
}

//APITideTeam holds the data expected from the cp api for this entity
type APITideTeam struct {
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	BucketUUID string `json:"bucket_uuid"`
}

//APICodeReference holds the data expected from the cp api for this entity
type APICodeReference struct {
	Branch         string      `json:"branch"`
	CodeRepository interface{} `json:"code_repository"`
	Sha1           string      `json:"sha1"`
}

//APIEnvironment holds the data expected from the cp api for this entity
type APIEnvironment struct {
	Cluster    string         `json:"cluster"`
	Components []APIComponent `json:"components"`
	Identifier string         `json:"identifier"`
}

//APIRepository holds the data expected from the cp api for this entity
type APIRepository struct {
	Address      string `json:"address"`
	Identifier   string `json:"identifier"`
	Name         string `json:"name"`
	Organisation string `json:"organisation"`
	Private      bool   `json:"private"`
	Type         string `json:"type"`
}

//APIFlow holds the data expected from the cp api for this entity
type APIFlow struct {
	UUID          string        `json:"uuid"`
	Configuration interface{}   `json:"configuration"`
	Pipelines     interface{}   `json:"pipelines"`
	Repository    APIRepository `json:"repository"`
	Team          APITeam       `json:"team"`
	Tides         []APITide     `json:"tides"`
	User          APIUser       `json:"user"`
}

//APIComponent holds the data expected from the cp api for this entity
type APIComponent struct {
	DeploymentStrategy interface{}   `json:"deployment_strategy"`
	Endpoints          []interface{} `json:"endpoints"`
	Extensions         []interface{} `json:"extensions"`
	Identifier         string        `json:"identifier"`
	Labels             interface{}   `json:"labels"`
	Name               string        `json:"name"`
	Specification      interface{}   `json:"specification"`
	Status             interface{}   `json:"status"`
}

//SetAPIKey sets the cp api key
func (c *CpAPI) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

//GetAPITeams sends a request to the cp api to fetch the list of teams for the user
func (c CpAPI) GetAPITeams() ([]APITeam, error) {
	if c.apiKey == "" {
		return nil, errors.Errorf(cperrors.StatusReasonFormat, http.StatusBadRequest, errorAPIKeyNotProvided)
	}

	u, err := c.getAuthenticatorURL()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToRetrievedAuthenticatorURL))
	}
	u.Path = "/api/teams"

	cplogs.V(5).Infof("getting api teams info on cp using url %s", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToCreateGetRequest))
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	respBody, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof(errorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToGetResponseBody))
	}

	teams := make([]APITeam, 0)
	err = json.Unmarshal(respBody, &teams)
	if err != nil {
		msg := fmt.Sprintf(errorParsingJSONResponse, respBody)
		cplogs.V(4).Infof(msg)
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusBadRequest, msg))
	}

	return teams, nil
}

//GetAPIFlows sends a request to the cp api to fetch the list of flows
func (c CpAPI) GetAPIFlows(project string) ([]APIFlow, error) {
	if c.apiKey == "" {
		return nil, errors.Errorf(cperrors.StatusReasonFormat, http.StatusBadRequest, errorAPIKeyNotProvided)
	}

	u, err := c.getRiverURL()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToRetrievedRiverURL))
	}
	u.Path = fmt.Sprintf("/teams/%s/flows", project)

	cplogs.V(5).Infof("getting api flows info on cp using url %s", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToCreateGetRequest))
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	respBody, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof(errorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToGetResponseBody))
	}

	flows := make([]APIFlow, 0)
	err = json.Unmarshal(respBody, &flows)
	if err != nil {
		msg := fmt.Sprintf(errorParsingJSONResponse, respBody)
		cplogs.V(4).Infof(msg)
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusBadRequest, msg))
	}

	return flows, nil
}

//GetAPIUser sends a request to the cp api to fetch the cp user information
func (c CpAPI) GetAPIUser(user string) (*APIUser, error) {
	if c.apiKey == "" {
		return nil, errors.Errorf(cperrors.StatusReasonFormat, http.StatusBadRequest, errorAPIKeyNotProvided)
	}

	u, err := c.getAuthenticatorURL()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToRetrievedAuthenticatorURL))
	}
	u.Path = "/api/user/" + user

	cplogs.V(5).Infof("getting user info on cp using url %s", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToCreateGetRequest))
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	respBody, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof(errorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToGetResponseBody))
	}

	apiUserResponse := &APIUser{}
	err = json.Unmarshal(respBody, apiUserResponse)
	if err != nil {
		msg := fmt.Sprintf(errorParsingJSONResponse, respBody)
		cplogs.V(4).Infof(msg)
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusBadRequest, msg))
	}

	return apiUserResponse, nil
}

//GetAPIEnvironments sends a request to the cp api to fetch the cp environments list for the user
func (c CpAPI) GetAPIEnvironments(flowID string) ([]APIEnvironment, error) {
	if c.apiKey == "" {
		return nil, errors.Errorf(cperrors.StatusReasonFormat, http.StatusBadRequest, errorAPIKeyNotProvided)
	}

	u, err := c.getRiverURL()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToRetrievedRiverURL))
	}
	u.Path = fmt.Sprintf("/flows/%s/environments", flowID)

	cplogs.V(5).Infof("getting flow environments using url %s", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToCreateGetRequest))
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	respBody, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof(errorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToGetResponseBody))
	}

	environments := make([]APIEnvironment, 0)
	err = json.Unmarshal(respBody, &environments)
	if err != nil {
		msg := fmt.Sprintf(errorParsingJSONResponse, respBody)
		cplogs.V(4).Infof(msg)
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusBadRequest, msg))
	}

	return environments, nil
}

//GetRemoteEnvironmentStatus sends a request to the cp api to retrieve information about the remote environment
func (c CpAPI) GetRemoteEnvironmentStatus(flowID string, environmentID string) (*APIRemoteEnvironmentStatus, error) {
	if c.apiKey == "" {
		return nil, errors.Errorf(cperrors.StatusReasonFormat, http.StatusBadRequest, errorAPIKeyNotProvided)
	}

	u, err := c.getRiverURL()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToRetrievedRiverURL))
	}
	u.Path = fmt.Sprintf("/flows/%s/development-environments/%s/status", flowID, environmentID)

	cplogs.V(5).Infof("getting remote environment status using url %s", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToCreateGetRequest))
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	respBody, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof(errorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToGetResponseBody))
	}

	apiRemoteEnvironment := &APIRemoteEnvironmentStatus{}
	err = json.Unmarshal(respBody, apiRemoteEnvironment)
	if err != nil {
		msg := fmt.Sprintf(errorParsingJSONResponse, respBody)
		cplogs.V(4).Infof(msg)
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusBadRequest, msg))
	}

	return apiRemoteEnvironment, nil
}

//RemoteEnvironmentBuild sends a request to the cp api to request to build a new remote environment
func (c CpAPI) RemoteEnvironmentBuild(remoteEnvironmentFlowID string, gitBranch string) error {
	if c.apiKey == "" {
		return errors.Errorf(cperrors.StatusReasonFormat, http.StatusBadRequest, errorAPIKeyNotProvided)
	}

	u, err := c.getRiverURL()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToRetrievedRiverURL))
	}
	u.Path = fmt.Sprintf("/flows/%s/tides", remoteEnvironmentFlowID)

	type requestBody struct {
		BranchName string `json:"branch"`
	}
	reqBodyJSON, err := json.Marshal(&requestBody{gitBranch})

	cplogs.V(5).Infof("triggering remote environment build using url %s and payload %s", u.Path, reqBodyJSON)

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(reqBodyJSON))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToCreatePostRequest))
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)

	_, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof(errorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToGetResponseBody))
	}

	return nil
}

//CancelRunningTide find the tide id associated with the branch and cancel the tide
func (c CpAPI) CancelRunningTide(flowID string, remoteEnvironmentID string) error {
	if c.apiKey == "" {
		return errors.Errorf(cperrors.StatusReasonFormat, http.StatusBadRequest, errorAPIKeyNotProvided)
	}

	remoteEnv, err := c.GetRemoteEnvironmentStatus(flowID, remoteEnvironmentID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToGetRemoteEnvironmentStatus))
	}

	if remoteEnv.LastTide.Status != TideRunning {
		cplogs.V(5).Infof("TideId %s not running, skipping", remoteEnv.LastTide.UUID)
		cplogs.Flush()
		return nil
	}

	return c.CancelTide(remoteEnv.LastTide.UUID)
}

//CancelTide send a request to the cp api to cancel the tide specified
func (c CpAPI) CancelTide(tideID string) error {
	if c.apiKey == "" {
		return errors.Errorf(cperrors.StatusReasonFormat, http.StatusBadRequest, errorAPIKeyNotProvided)
	}

	u, err := c.getRiverURL()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToRetrievedRiverURL))
	}
	u.Path = fmt.Sprintf("/tides/%s/cancel", tideID)

	cplogs.V(5).Infof("cancelling tide using url %s", u.String())

	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToCreatePostRequest))
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)

	_, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof(errorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToGetResponseBody))
	}

	return nil
}

//RemoteEnvironmentDestroy sends a request to the cp api to request to destroy the remote environment
func (c CpAPI) RemoteEnvironmentDestroy(flowID string, environment string, cluster string) error {
	if c.apiKey == "" {
		return errors.Errorf(cperrors.StatusReasonFormat, http.StatusBadRequest, errorAPIKeyNotProvided)
	}

	u, err := c.getRiverURL()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToRetrievedRiverURL))
	}
	u.Path = fmt.Sprintf("/flows/%s/environments/%s", flowID, environment)
	u.RawQuery = fmt.Sprintf("cluster=%s", cluster)

	cplogs.V(5).Infof("destroying remote environment using url %s", u.String())

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToCreateDeleteRequest))
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)

	_, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof(errorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToGetResponseBody))
	}

	return nil
}

//RemoteEnvironmentRunningAndExists get the remote environment status and the list of all environments to ensure that the environment actually exists
func (c CpAPI) RemoteEnvironmentRunningAndExists(flowID string, environmentID string) (bool, error) {
	if c.apiKey == "" {
		return false, errors.Errorf(cperrors.StatusReasonFormat, http.StatusBadRequest, errorAPIKeyNotProvided)
	}

	remoteEnv, err := c.GetRemoteEnvironmentStatus(flowID, environmentID)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToGetRemoteEnvironmentStatus))
	}

	environments, err := c.GetAPIEnvironments(flowID)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToGetEnvironmentsList))
	}
	for _, environment := range environments {
		if environment.Identifier == remoteEnv.KubeEnvironmentName {
			return true, nil
		}
	}
	return true, nil
}

//RemoteDevelopmentEnvironmentDestroy send a request to the cp api to destroy the environment
func (c CpAPI) RemoteDevelopmentEnvironmentDestroy(flowID string, remoteEnvironmentID string) error {
	if c.apiKey == "" {
		return errors.Errorf(cperrors.StatusReasonFormat, http.StatusBadRequest, errorAPIKeyNotProvided)
	}

	u, err := c.getRiverURL()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToRetrievedRiverURL))
	}
	u.Path = fmt.Sprintf("/flows/%s/development-environments/%s", flowID, remoteEnvironmentID)

	cplogs.V(5).Infof("destroying remote development environment using url %s", u.String())

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToCreateDeleteRequest))
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)

	_, elr := c.getResponseBody(c.client, req)
	if elr != nil {
		cplogs.V(4).Infof(errorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errorFailedToGetResponseBody))
	}

	return nil
}

func (c CpAPI) getResponseBody(client *http.Client, req *http.Request) ([]byte, error) {
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, err))
	}
	defer res.Body.Close()
	if res.StatusCode < 200 && res.StatusCode > 202 {
		errStr := fmt.Sprintf(errorResponseStatusCodeUnsuccessful, res.StatusCode, req.URL.String())
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, errStr))
	}
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, err))
	}
	return resBody, nil
}

func (c CpAPI) getAuthenticatorURL() (*url.URL, error) {
	cpAPIAddr, err := config.C.GetString(config.CpAuthenticatorApiAddr)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, err))
	}
	u, err := url.Parse(cpAPIAddr)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, err))
	}
	return u, nil
}

func (c CpAPI) getRiverURL() (*url.URL, error) {
	cpAPIAddr, err := config.C.GetString(config.CpRiverApiAddr)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, err))
	}
	u, err := url.Parse(cpAPIAddr)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(cperrors.StatusReasonFormat, http.StatusInternalServerError, err))
	}
	return u, nil
}

//PrintPublicEndpoints given a list of public api endpoints it prints them on the given writer
func PrintPublicEndpoints(writer io.Writer, endpoints []APIPublicEndpoint) {
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
