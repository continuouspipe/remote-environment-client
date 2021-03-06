//logs data on keen.io
package keenapi

import (
	"bufio"
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

	cplogs.V(5).Infof("sending POST request with payload %s", reader)
	cplogs.V(7).Infof("sending POST request to %s, payload %s", k.getEndpointUrl(), reader)
	req, err := http.NewRequest("POST", k.getEndpointUrl(), reader)
	if err != nil {
		cplogs.V(4).Infof("could not create request for GET request for url: %s", k.getEndpointUrl())
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		cplogs.V(4).Infof("could not execute the GET request for url: %s", k.getEndpointUrl())
		return false, err
	}
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		return true, nil
	}

	respBody := ""
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		respBody = respBody + "\n" + scanner.Text()
	}

	err = fmt.Errorf("we didn't receive the expected status code OK or Create from keen.io. Status code %d, body %s", resp.StatusCode, respBody)
	return false, err
}

func (k *Sender) getEndpointUrl() string {
	return fmt.Sprintf("https://api.keen.io/3.0/projects/%s/events/%s?api_key=%s",
		k.ProjectId,
		k.EventCollection,
		k.WriteKey)
}
