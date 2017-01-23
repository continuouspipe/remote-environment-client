package keenapi

import (
	"encoding/json"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"time"
)

type BenchmarkPayload struct {
	settings  config.Reader
	Command   string
	StartTime string
}

func NewBenchmarkPayload() *BenchmarkPayload {
	b := &BenchmarkPayload{}
	b.settings = config.NewApplicationSettings()
	return b
}

func (b *BenchmarkPayload) GetJsonPayload() ([]byte, error) {
	t := time.Now()
	endTime := t.Format(time.RFC3339)

	payload := make(map[string]string)
	payload["project"] = b.settings.GetString(config.ProjectKey)
	payload["namespace"] = b.settings.GetString(config.Environment)
	payload["command"] = b.settings.GetString(b.Command)
	payload["start-time"] = b.StartTime
	payload["end-time"] = endTime
	out, err := json.Marshal(payload)
	if err != nil {
		cplogs.V(4).Infof("could not generate the benchmark json payload for %#v", payload)
		return nil, err
	}
	return out, nil
}
