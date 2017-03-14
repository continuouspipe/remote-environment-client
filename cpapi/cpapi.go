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
	"io"
)

type CpApiProvider interface {
	SetApiKey(apiKey string)
	GetApiTeams() ([]ApiTeam, error)
	GetApiBucketClusters(bucketUuid string) ([]ApiCluster, error)
	GetApiUser(user string) (*ApiUser, error)
	GetApiEnvironments(flowId string) ([]ApiEnvironment, error)
	GetRemoteEnvironmentStatus(flowId string, environmentId string) (*ApiRemoteEnvironmentStatus, error)
	RemoteEnvironmentBuild(remoteEnvironmentFlowID string, gitBranch string) error
	CancelRunningTide(flowId string, remoteEnvironmentId string) error
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
	if c.apiKey == "" {
		return nil, fmt.Errorf("Api key not provided.")
	}

	url, err := c.getAuthenticatorURL()
	if err != nil {
		return nil, err
	}
	url.Path = "/api/teams"

	cplogs.V(5).Infof("getting api teams info on cp using url %s", url.String())

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	req.Header.Add("X-Api-Key", c.apiKey)
	if err != nil {
		return nil, err
	}

	respBody, err := c.getResponseBody(c.client, req)

	if err != nil {
		return nil, err
	}

	teams := make([]ApiTeam, 0)
	err = json.Unmarshal(respBody, teams)
	if err != nil {
		cplogs.V(4).Infof("error running Unmarshal() on response body %s", respBody)
		return nil, err
	}

	return teams, nil
}

//Use the master api key to get the details of the cluster, including the auth password for kubernetes in cleartext
func (c CpApi) GetApiBucketClusters(bucketUuid string) ([]ApiCluster, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("Api key not provided.")
	}

	url, err := c.getAuthenticatorURL()
	if err != nil {
		return nil, err
	}
	url.Path = "/api/bucket/" + bucketUuid + "/clusters"

	cplogs.V(5).Infof("getting api bucke cluster info on cp using url %s", url.String())

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)

	req.Header.Add("X-Api-Key", c.apiKey)
	if err != nil {
		return nil, err
	}

	respBody, err := c.getResponseBody(c.client, req)
	if err != nil {
		return nil, err
	}

	clusters := make([]ApiCluster, 0)
	err = json.Unmarshal(respBody, &clusters)
	if err != nil {
		cplogs.V(4).Infof("error running Unmarshal() on response body %s", respBody)
		return nil, err
	}

	return clusters, nil
}

func (c CpApi) GetApiUser(user string) (*ApiUser, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("Api key not provided.")
	}

	url, err := c.getAuthenticatorURL()
	if err != nil {
		return nil, err
	}
	url.Path = "/api/user/" + user

	cplogs.V(5).Infof("getting user info on cp using url %s", url.String())

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	req.Header.Add("X-Api-Key", c.apiKey)
	if err != nil {
		return nil, err
	}

	respBody, err := c.getResponseBody(c.client, req)
	if err != nil {
		return nil, err
	}

	apiUserResponse := &ApiUser{}
	err = json.Unmarshal(respBody, apiUserResponse)
	if err != nil {
		cplogs.V(4).Infof("error running Unmarshal() on response body %s", respBody)
		return nil, err
	}

	return apiUserResponse, nil
}

func (c CpApi) GetApiEnvironments(flowId string) ([]ApiEnvironment, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("api key not provided")
	}

	url, err := c.getRiverURL()
	if err != nil {
		return nil, err
	}
	url.Path = fmt.Sprintf("/flows/%s/environments", flowId)

	cplogs.V(5).Infof("getting flow environments using url %s", url.String())

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	req.Header.Add("X-Api-Key", c.apiKey)
	if err != nil {
		return nil, err
	}

	respBody, err := c.getResponseBody(c.client, req)
	if err != nil {
		return nil, fmt.Errorf("error getting remote environment, %s", err.Error())
	}

	environments := make([]ApiEnvironment, 0)
	err = json.Unmarshal(respBody, &environments)
	if err != nil {
		cplogs.V(4).Infof("error running Unmarshal() on response body %s", respBody)
		return nil, err
	}

	return environments, nil
}

//calls CP Api to retrieve information about the remote environment
func (c CpApi) GetRemoteEnvironmentStatus(flowId string, environmentId string) (*ApiRemoteEnvironmentStatus, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("api key not provided")
	}

	url, err := c.getRiverURL()
	if err != nil {
		return nil, err
	}
	url.Path = fmt.Sprintf("/flows/%s/development-environments/%s/status", flowId, environmentId)

	cplogs.V(5).Infof("getting remote environment status using url %s", url.String())

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	req.Header.Add("X-Api-Key", c.apiKey)
	if err != nil {
		return nil, err
	}

	respBody, err := c.getResponseBody(c.client, req)
	if err != nil {
		return nil, fmt.Errorf("error getting remote environment, %s", err.Error())
	}

	apiRemoteEnvironment := &ApiRemoteEnvironmentStatus{}
	err = json.Unmarshal(respBody, apiRemoteEnvironment)
	if err != nil {
		cplogs.V(4).Infof("error running Unmarshal() on response body %s", respBody)
		return nil, err
	}

	return apiRemoteEnvironment, nil
}

//calls CP API to request to build a new remote environment
func (c CpApi) RemoteEnvironmentBuild(remoteEnvironmentFlowID string, gitBranch string) error {
	if c.apiKey == "" {
		return fmt.Errorf("api key not provided")
	}

	url, err := c.getRiverURL()
	if err != nil {
		return err
	}
	url.Path = fmt.Sprintf("/flows/%s/tides", remoteEnvironmentFlowID)

	type requestBody struct {
		BranchName string `json:"branch"`
	}
	reqBodyJson, err := json.Marshal(&requestBody{gitBranch})

	cplogs.V(5).Infof("triggering remote environment build using url %s and payload %s", url.Path, reqBodyJson)

	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewReader(reqBodyJson))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)
	if err != nil {
		return err
	}

	_, err = c.getResponseBody(c.client, req)
	if err != nil {
		return err
	}

	return nil
}

//find the tide id associated with the branch and cancel the tide
func (c CpApi) CancelRunningTide(flowId string, remoteEnvironmentId string) error {
	remoteEnv, err := c.GetRemoteEnvironmentStatus(flowId, remoteEnvironmentId)
	if err != nil {
		return err
	}

	if remoteEnv.LastTide.Status != TideRunning {
		cplogs.V(5).Infof("TideId %s not running, skipping", remoteEnv.LastTide.Uuid)
		return nil
	}

	return c.CancelTide(remoteEnv.LastTide.Uuid)
}

func (c CpApi) CancelTide(tideId string) error {
	if c.apiKey == "" {
		return fmt.Errorf("api key not provided")
	}

	url, err := c.getRiverURL()
	if err != nil {
		return err
	}
	url.Path = fmt.Sprintf("/tides/%s/cancel", tideId)

	cplogs.V(5).Infof("cancelling tide using url %s", url.String())

	req, err := http.NewRequest(http.MethodPost, url.String(), nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)
	if err != nil {
		return err
	}

	_, err = c.getResponseBody(c.client, req)
	if err != nil {
		return err
	}

	return nil
}

//calls CP API to request to destroy the remote environment
func (c CpApi) RemoteEnvironmentDestroy(flowId string, environment string, cluster string) error {
	if c.apiKey == "" {
		return fmt.Errorf("api key not provided")
	}

	url, err := c.getRiverURL()
	if err != nil {
		return err
	}
	url.Path = fmt.Sprintf("/flows/%s/environments/%s", flowId, environment)
	url.RawQuery = fmt.Sprintf("cluster=%s", cluster)

	cplogs.V(5).Infof("destroying remote environment using url %s", url.String())

	req, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)
	if err != nil {
		return err
	}

	_, err = c.getResponseBody(c.client, req)
	if err != nil {
		return err
	}

	return nil
}

func (c CpApi) RemoteDevelopmentEnvironmentDestroy(flowId string, remoteEnvironmentId string) error {
	if c.apiKey == "" {
		return fmt.Errorf("api key not provided")
	}

	url, err := c.getRiverURL()
	if err != nil {
		return err
	}
	url.Path = fmt.Sprintf("/flows/%s/development-environments/%s", flowId, remoteEnvironmentId)

	cplogs.V(5).Infof("destroying remote development environment using url %s", url.String())

	req, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Api-Key", c.apiKey)
	if err != nil {
		return err
	}

	_, err = c.getResponseBody(c.client, req)
	if err != nil {
		return err
	}

	return nil
}

func (c CpApi) getResponseBody(client *http.Client, req *http.Request) ([]byte, error) {
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 && res.StatusCode > 202 {
		return nil, fmt.Errorf("error getting response body, status: %d, url: %s", res.StatusCode, req.URL.String())
	}
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
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
