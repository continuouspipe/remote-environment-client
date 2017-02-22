package keenapi

import (
	"encoding/json"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"time"
)

type BenchmarkPayload struct {
	settings  *config.Config
	Command   string
	StartTime string
}

func NewBenchmarkPayload() *BenchmarkPayload {
	b := &BenchmarkPayload{}
	b.settings = config.C
	return b
}

func (b *BenchmarkPayload) GetJsonPayload() ([]byte, error) {
	t := time.Now()
	endTime := t.Format(time.RFC3339)

	project, err := b.settings.GetString(config.Project)
	if err != nil {
		return nil, err
	}
	environment, err := b.settings.GetString(config.KubeEnvironmentName)
	if err != nil {
		return nil, err
	}

	payload := make(map[string]string)
	payload["project"] = project
	payload["namespace"] = environment
	payload["command"] = b.Command
	payload["start-time"] = b.StartTime
	payload["end-time"] = endTime
	out, err := json.Marshal(payload)
	if err != nil {
		cplogs.V(4).Infof("could not generate the benchmark json payload for %#v", payload)
		return nil, err
	}
	return out, nil
}
