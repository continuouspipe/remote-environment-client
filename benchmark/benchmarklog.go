package benchmark

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/continuouspipe/remote-environment-client/keenapi"
	"time"
)

type CmdBenchmark struct {
	sender          *keenapi.Sender
	payloadProvider *keenapi.BenchmarkPayload
	currentCommand  string
	startTime       string
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
	b.currentCommand = cmd
	b.startTime = time.Now().Format(time.RFC3339)
}

func (b *CmdBenchmark) StopAndLog(cmd string) (bool, error) {
	if cmd != b.currentCommand {
		return false, fmt.Errorf("stop the benchmark in progress %s or start a new one for %s", b.currentCommand, cmd)
	}

	res, err := b.sender.Send(b.payloadProvider)
	if err != nil {
		cplogs.Errorf("an error occured when stopping the command %s, %#s", cmd, err.Error())
		return false, err
	}

	b.currentCommand = ""
	b.startTime = ""
	return res, nil
}
