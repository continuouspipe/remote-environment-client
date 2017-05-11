package cpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"net/url"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	cperrors "github.com/continuouspipe/remote-environment-client/errors"
	cphttp "github.com/continuouspipe/remote-environment-client/http"
	"github.com/pkg/errors"
)

const errorAPIKeyNotProvided = "api key not provided"
const errorFailedToRetrievedAuthenticatorURL = "failed to retrieve the authenticator url"
const errorFailedToRetrievedRiverURL = "failed to retrieve the river url"
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
		return nil, errors.New(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, errorAPIKeyNotProvided).String())
	}

	u, err := GetAuthenticatorURL()
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToRetrievedAuthenticatorURL).String())
	}
	u.Path = "/api/teams"

	cplogs.V(5).Infof("getting api teams info on cp using url %s", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToCreateGetRequest).String())
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	respBody, err := cphttp.GetResponseBody(c.client, req)
	if err != nil {
		cplogs.V(4).Infof(cphttp.ErrorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToGetResponseBody).String())
	}

	teams := make([]APITeam, 0)
	err = json.Unmarshal(respBody, &teams)
	if err != nil {
		msg := fmt.Sprintf(cphttp.ErrorParsingJSONResponse, respBody)
		cplogs.V(4).Infof(msg)
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msg).String())
	}

	return teams, nil
}

//GetAPIFlows sends a request to the cp api to fetch the list of flows
func (c CpAPI) GetAPIFlows(project string) ([]APIFlow, error) {
	if c.apiKey == "" {
		return nil, errors.Errorf(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, errorAPIKeyNotProvided).String())
	}

	u, err := GetRiverURL()
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToRetrievedRiverURL).String())
	}
	u.Path = fmt.Sprintf("/teams/%s/flows", project)

	cplogs.V(5).Infof("getting api flows info on cp using url %s", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToCreateGetRequest).String())
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	respBody, err := cphttp.GetResponseBody(c.client, req)
	if err != nil {
		cplogs.V(4).Infof(cphttp.ErrorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToGetResponseBody).String())
	}

	flows := make([]APIFlow, 0)
	err = json.Unmarshal(respBody, &flows)
	if err != nil {
		msg := fmt.Sprintf(cphttp.ErrorParsingJSONResponse, respBody)
		cplogs.V(4).Infof(msg)
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msg).String())
	}

	return flows, nil
}

//GetAPIUser sends a request to the cp api to fetch the cp user information
func (c CpAPI) GetAPIUser(user string) (*APIUser, error) {
	if c.apiKey == "" {
		return nil, errors.Errorf(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, errorAPIKeyNotProvided).String())
	}

	u, err := GetAuthenticatorURL()
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToRetrievedAuthenticatorURL).String())
	}
	u.Path = "/api/user/" + user

	cplogs.V(5).Infof("getting user info on cp using url %s", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToCreateGetRequest).String())
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	respBody, err := cphttp.GetResponseBody(c.client, req)
	if err != nil {
		cplogs.V(4).Infof(cphttp.ErrorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToGetResponseBody).String())
	}

	apiUserResponse := &APIUser{}
	err = json.Unmarshal(respBody, apiUserResponse)
	if err != nil {
		msg := fmt.Sprintf(cphttp.ErrorParsingJSONResponse, respBody)
		cplogs.V(4).Infof(msg)
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msg).String())
	}

	return apiUserResponse, nil
}

//GetAPIEnvironments sends a request to the cp api to fetch the cp environments list for the user
func (c CpAPI) GetAPIEnvironments(flowID string) ([]APIEnvironment, error) {
	if c.apiKey == "" {
		return nil, errors.Errorf(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, errorAPIKeyNotProvided).String())
	}

	u, err := GetRiverURL()
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToRetrievedRiverURL).String())
	}
	u.Path = fmt.Sprintf("/flows/%s/environments", flowID)

	cplogs.V(5).Infof("getting flow environments using url %s", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToCreateGetRequest).String())
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	respBody, err := cphttp.GetResponseBody(c.client, req)
	if err != nil {
		cplogs.V(4).Infof(cphttp.ErrorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToGetResponseBody).String())
	}

	environments := make([]APIEnvironment, 0)
	err = json.Unmarshal(respBody, &environments)
	if err != nil {
		msg := fmt.Sprintf(cphttp.ErrorParsingJSONResponse, respBody)
		cplogs.V(4).Infof(msg)
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msg).String())
	}

	return environments, nil
}

//GetRemoteEnvironmentStatus sends a request to the cp api to retrieve information about the remote environment
func (c CpAPI) GetRemoteEnvironmentStatus(flowID string, environmentID string) (*APIRemoteEnvironmentStatus, error) {
	if c.apiKey == "" {
		return nil, errors.Errorf(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, errorAPIKeyNotProvided).String())
	}

	u, err := GetRiverURL()
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToRetrievedRiverURL).String())
	}
	u.Path = fmt.Sprintf("/flows/%s/development-environments/%s/status", flowID, environmentID)

	cplogs.V(5).Infof("getting remote environment status using url %s", u.String())

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToCreateGetRequest).String())
	}
	req.Header.Add("X-Api-Key", c.apiKey)

	respBody, err := cphttp.GetResponseBody(c.client, req)
	if err != nil {
		cplogs.V(4).Infof(cphttp.ErrorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToGetResponseBody).String())
	}

	apiRemoteEnvironment := &APIRemoteEnvironmentStatus{}
	err = json.Unmarshal(respBody, apiRemoteEnvironment)
	if err != nil {
		msg := fmt.Sprintf(cphttp.ErrorParsingJSONResponse, respBody)
		cplogs.V(4).Infof(msg)
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusBadRequest, msg).String())
	}

	return apiRemoteEnvironment, nil
}

//RemoteEnvironmentBuild sends a request to the cp api to request to build a new remote environment
func (c CpAPI) RemoteEnvironmentBuild(remoteEnvironmentFlowID string, gitBranch string) error {
	if c.apiKey == "" {
		return errors.Errorf(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, errorAPIKeyNotProvided).String())
	}

	u, err := GetRiverURL()
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToRetrievedRiverURL).String())
	}
	u.Path = fmt.Sprintf("/flows/%s/tides", remoteEnvironmentFlowID)

	type requestBody struct {
		BranchName string `json:"branch"`
	}
	reqBodyJSON, err := json.Marshal(&requestBody{gitBranch})
	if err != nil {
		msg := fmt.Sprintf(cphttp.ErrorCreatingJSONRequest, &requestBody{gitBranch})
		cplogs.V(4).Infof(msg)
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, msg).String())
	}

	cplogs.V(5).Infof("triggering remote environment build using url %s and payload %s", u.Path, reqBodyJSON)

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(reqBodyJSON))
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToCreatePostRequest).String())
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)

	_, err = cphttp.GetResponseBody(c.client, req)
	if err != nil {
		cplogs.V(4).Infof(cphttp.ErrorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToGetResponseBody).String())
	}

	return nil
}

//CancelRunningTide find the tide id associated with the branch and cancel the tide
func (c CpAPI) CancelRunningTide(flowID string, remoteEnvironmentID string) error {
	if c.apiKey == "" {
		return errors.Errorf(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, errorAPIKeyNotProvided).String())
	}

	remoteEnv, err := c.GetRemoteEnvironmentStatus(flowID, remoteEnvironmentID)
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToGetRemoteEnvironmentStatus).String())
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
		return errors.Errorf(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, errorAPIKeyNotProvided).String())
	}

	u, err := GetRiverURL()
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToRetrievedRiverURL).String())
	}
	u.Path = fmt.Sprintf("/tides/%s/cancel", tideID)

	cplogs.V(5).Infof("cancelling tide using url %s", u.String())

	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToCreatePostRequest).String())
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)

	_, err = cphttp.GetResponseBody(c.client, req)
	if err != nil {
		cplogs.V(4).Infof(cphttp.ErrorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToGetResponseBody).String())
	}

	return nil
}

//RemoteEnvironmentDestroy sends a request to the cp api to request to destroy the remote environment
func (c CpAPI) RemoteEnvironmentDestroy(flowID string, environment string, cluster string) error {
	if c.apiKey == "" {
		return errors.Errorf(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, errorAPIKeyNotProvided).String())
	}

	u, err := GetRiverURL()
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToRetrievedRiverURL).String())
	}
	u.Path = fmt.Sprintf("/flows/%s/environments/%s", flowID, environment)
	u.RawQuery = fmt.Sprintf("cluster=%s", cluster)

	cplogs.V(5).Infof("destroying remote environment using url %s", u.String())

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToCreateDeleteRequest).String())
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)

	_, err = cphttp.GetResponseBody(c.client, req)
	if err != nil {
		cplogs.V(4).Infof(cphttp.ErrorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToGetResponseBody).String())
	}

	return nil
}

//RemoteEnvironmentRunningAndExists get the remote environment status and the list of all environments to ensure that the environment actually exists
func (c CpAPI) RemoteEnvironmentRunningAndExists(flowID string, environmentID string) (bool, error) {
	if c.apiKey == "" {
		return false, errors.Errorf(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, errorAPIKeyNotProvided).String())
	}

	remoteEnv, err := c.GetRemoteEnvironmentStatus(flowID, environmentID)
	if err != nil {
		return false, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToGetRemoteEnvironmentStatus).String())
	}

	environments, err := c.GetAPIEnvironments(flowID)
	if err != nil {
		return false, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToGetEnvironmentsList).String())
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
		return errors.Errorf(cperrors.NewStatefulErrorMessage(http.StatusBadRequest, errorAPIKeyNotProvided).String())
	}

	u, err := GetRiverURL()
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, errorFailedToRetrievedRiverURL).String())
	}
	u.Path = fmt.Sprintf("/flows/%s/development-environments/%s", flowID, remoteEnvironmentID)

	cplogs.V(5).Infof("destroying remote development environment using url %s", u.String())

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToCreateDeleteRequest).String())
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)

	_, err = cphttp.GetResponseBody(c.client, req)
	if err != nil {
		cplogs.V(4).Infof(cphttp.ErrorFailedToGetResponseBody, u.String())
		cplogs.Flush()
		return errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, cphttp.ErrorFailedToGetResponseBody).String())
	}

	return nil
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

func GetAuthenticatorURL() (*url.URL, error) {
	cpAPIAddr, err := config.C.GetString(config.CpAuthenticatorApiAddr)
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, err.Error()).String())
	}
	u, err := url.Parse(cpAPIAddr)
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, err.Error()).String())
	}
	return u, nil
}

func GetRiverURL() (*url.URL, error) {
	cpAPIAddr, err := config.C.GetString(config.CpRiverApiAddr)
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, err.Error()).String())
	}
	u, err := url.Parse(cpAPIAddr)
	if err != nil {
		return nil, errors.Wrap(err, cperrors.NewStatefulErrorMessage(http.StatusInternalServerError, err.Error()).String())
	}
	return u, nil
}
