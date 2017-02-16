package cpapi

import (
	"net/url"
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
)

type CpApiProvider interface {
	SetApiKey(apiKey string)
	GetApiTeams() ([]ApiTeam, error)
	GetApiBucketClusters(bucketUuid string) ([]ApiCluster, error)
	GetApiUser(user string) (*ApiUser, error)
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

type ApiTeam struct {
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	BucketUuid string      `json:"bucket_uuid"`

	//Should be []ApiMembership although there is a bug on the api where a list of object with keys "1", "2" is returned
	//instead of being a json array
	Memberships []interface{} `json:"memberships"`
}

type ApiMembership struct {
	Team        ApiTeam `json:"team"`
	User        ApiUser `json:"user"`
	Permissions []string `json:"permissions"`
}

type ApiUser struct {
	Username   string      `json:"username"`
	Email      string      `json:"email"`
	BucketUuid string      `json:"bucket_uuid"`
	Roles      []string      `json:"roles"`
}

type ApiCluster struct {
	Identifier string    `json:"identifier"`
	Address    string    `json:"address"`
	Version    string    `json:"version"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	Type       string    `json:"type"`
}

func (c *CpApi) SetApiKey(apiKey string) {
	c.apiKey = apiKey
}

func (c CpApi) GetApiTeams() ([]ApiTeam, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("Api key not provided.")
	}

	url := c.getAuthenticatorUrl()
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

	url := c.getAuthenticatorUrl()
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
	
	url := c.getAuthenticatorUrl()
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

func (c CpApi) getResponseBody(client *http.Client, req *http.Request) ([]byte, error) {
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error requesting user information, request status: %s", res.StatusCode)
	}
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func (c CpApi) getAuthenticatorUrl() *url.URL {
	//TODO: Read this from global config
	return &url.URL{
		Scheme: "https",
		Host:   "authenticator-staging.continuouspipe.io",
	}
}
