package cpapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"bytes"
	"github.com/continuouspipe/remote-environment-client/config"
)

type CpApiProvider interface {
	SetApiKey(apiKey string)
	GetApiTeams() ([]ApiTeam, error)
	GetApiBucketClusters(bucketUuid string) ([]ApiCluster, error)
	GetApiUser(user string) (*ApiUser, error)
	GetRemoteEnvironment(remoteEnvironmentID string) (*ApiRemoteEnvironment, error)
	RemoteEnvironmentBuild(remoteEnvironmentID string, gitBranch string) error
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

const RemoteEnvironmentStatusOk = "Ok"
const RemoteEnvironmentStatusFailed = "TideFailed"
const RemoteEnvironmentStatusBuilding = "TideRunning"
const RemoteEnvironmentStatusNotStarted = "NotStarted"

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

type ApiRemoteEnvironment struct {
	Status              string `json:"status"`
	ModifiedAt          string `json:"modified_at"`
	RemoteEnvironmentId string `json:"remote_environment_id"`
	KubeEnvironmentName string `json:"kubernetes_environment_name"`
	ClusterIdentifier   string `json:"cluster_identifier"`
	AnyBarPort          string `json:"any_bar_port"`
	KeenId              string `json:"keen_id"`
	KeenWriteKey        string `json:"keen_write_key"`
	KeenEventCollection string `json:"keen_event_collection"`
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

	req, err := http.NewRequest("GET", url.String(), nil)
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

	req, err := http.NewRequest("GET", url.String(), nil)

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

	req, err := http.NewRequest("GET", url.String(), nil)
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
		return nil, err
	}

	return apiUserResponse, nil
}

//GetRemoteEnvironment call CP Api to retrieve information about the remote environment
func (c CpApi) GetRemoteEnvironment(remoteEnvironmentID string) (*ApiRemoteEnvironment, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("api key not provided")
	}

	url, err := c.getAuthenticatorURL()
	if err != nil {
		return nil, err
	}
	url.Path = "/api/remote-environment/" + remoteEnvironmentID

	req, err := http.NewRequest("GET", url.String(), nil)
	req.Header.Add("X-Api-Key", c.apiKey)
	if err != nil {
		return nil, err
	}

	respBody, err := c.getResponseBody(c.client, req)
	if err != nil {
		return nil, fmt.Errorf("error getting remote environment, %s", err.Error())
	}

	apiRemoteEnvironment := &ApiRemoteEnvironment{}
	err = json.Unmarshal(respBody, apiRemoteEnvironment)
	if err != nil {
		return nil, err
	}

	return apiRemoteEnvironment, nil
}

//RemoteEnvironmentBuild call CP API to request to build a new remote environment
func (c CpApi) RemoteEnvironmentBuild(remoteEnvironmentID string, gitBranch string) error {
	if c.apiKey == "" {
		return fmt.Errorf("api key not provided")
	}

	url, err := c.getAuthenticatorURL()
	if err != nil {
		return err
	}
	url.Path = "/api/remote-environment/" + remoteEnvironmentID + "/build"

	type requestBody struct {
		BranchName string `json:"branch_name"`
	}
	reqBodyJson, err := json.Marshal(&requestBody{gitBranch})

	req, err := http.NewRequest("POST", url.String(), bytes.NewReader(reqBodyJson))
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
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting response body, status: %d, url: %s", res.StatusCode, req.URL.String())
	}
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func (c CpApi) getAuthenticatorURL() (*url.URL, error) {
	cpApiAddr, err := config.C.GetString(config.CpApiAddr)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(cpApiAddr)
	if err != nil {
		return nil, err
	}
	return u, nil
}
