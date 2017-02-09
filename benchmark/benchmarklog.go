package benchmark

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/keenapi"
	"time"
)

type CmdBenchmark struct {
	sender          *keenapi.Sender
	payloadProvider *keenapi.BenchmarkPayload
}

func NewCmdBenchmark() *CmdBenchmark {
	settings := config.NewApplicationSettings()
	b := &CmdBenchmark{}
	b.sender = keenapi.NewSender()
	b.sender.ProjectId = settings.GetString(config.KeenProjectId)
	b.sender.WriteKey = settings.GetString(config.KeenWriteKey)
	b.sender.EventCollection = settings.GetString(config.KeenEventCollection)
	b.payloadProvider = keenapi.NewBenchmarkPayload()
	return b
}

func (b *CmdBenchmark) Start(cmd string) {
	if !b.keenSettingsAvailable() {
		cplogs.V(5).Infoln("keen.io settings disabled")
		return
	}

	b.payloadProvider.StartTime = time.Now().Format(time.RFC3339)
	b.payloadProvider.Command = cmd

	cplogs.V(5).Infof("started benchmarking %s at %s", b.payloadProvider.Command, b.payloadProvider.StartTime)
}

func (b *CmdBenchmark) StopAndLog() (bool, error) {
	if !b.keenSettingsAvailable() {
		cplogs.V(5).Infoln("keen.io settings disabled")
		return false, nil
	}

	res, err := b.sender.Send(b.payloadProvider)
	if err != nil {
		cplogs.V(4).Infof("an error occured when stopping the command %s, %#s", b.payloadProvider.Command, err.Error())
		return false, err
	}

	b.payloadProvider.Command = ""
	b.payloadProvider.StartTime = ""
	return res, nil
}

func (b *CmdBenchmark) keenSettingsAvailable() bool {
	return len(b.sender.ProjectId) > 0 && len(b.sender.WriteKey) > 0 && len(b.sender.EventCollection) > 0
}
