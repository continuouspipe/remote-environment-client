package keenapi

import (
	"encoding/json"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"time"
)

type BenchmarkPayload struct {
	settings  config.Reader
	command   string
	startTime string
}

func NewBenchmarkPayload() *BenchmarkPayload {
	return &BenchmarkPayload{}
}

func (b *BenchmarkPayload) GetJsonPayload() ([]byte, error) {
	t := time.Now()
	endTime := t.Format(time.RFC3339)

	payload := make(map[string]string)
	payload["project"] = b.settings.GetString(config.ProjectKey)
	payload["namespace"] = b.settings.GetString(config.Environment)
	payload["command"] = b.settings.GetString(b.command)
	payload["start-time"] = b.settings.GetString(b.startTime)
	payload["end-time"] = b.settings.GetString(endTime)
	out, err := json.Marshal(payload)
	if err != nil {
		cplogs.Errorf("could not generate the benchmark json payload for %#v", payload)
		return nil, err
	}
	return out, nil
}
