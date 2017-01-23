//logs data on keen.io
package keenapi

import (
	"bytes"
	"fmt"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"net/http"
)

type PayloadProvider interface {
	GetJsonPayload() ([]byte, error)
}

type PayloadSender interface {
	Send(payload PayloadProvider) (bool, error)
}

type Sender struct {
	ProjectId, EventCollection, WriteKey string
}

func NewSender() *Sender {
	return &Sender{}
}

func (k *Sender) Send(payload PayloadProvider) (bool, error) {
	out, err := payload.GetJsonPayload()
	if err != nil {
		return false, err
	}
	reader := bytes.NewReader(out)
	req, err := http.NewRequest("GET", k.getEndpointUrl(), reader)
	if err != nil {
		cplogs.Errorf("could not create request for GET request for url: %s", k.getEndpointUrl())
		return false, err
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		cplogs.Errorf("could not execute the GET request for url: %s", k.getEndpointUrl())
		return false, err
	}
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		return true, nil
	}
	err = fmt.Errorf("we didn't receive the expected status code OK or Create from keen.io. Status code %d, body %s", resp.StatusCode, resp.Body)
	return false, err
}

func (k *Sender) getEndpointUrl() string {
	return fmt.Sprintf("https://api.keen.io/3.0/projects/%s/events/%s?api_key=%s",
		k.ProjectId,
		k.EventCollection,
		k.WriteKey)
}
