package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/benchmark"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/cpapi"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var (
	startTideExample = fmt.Sprintf(`
		# Start a new tide
		%[1]s start --interactive ([-i]) <flow-uuid> <branch-or-commit>

		e.g.: %[1]s start --interactive 1268cc54-b265-11e6-b835-0c360641bb54 cpdev/someuser

		# Start a new tide on an initialised environment
		%[1]s start <flow-uuid> <branch-or-commit>
		`, config.AppName)
)

func NewTideStartCmd() *cobra.Command {
	handler := &TideStartHandle{}
	handler.stdout = os.Stdout
	handler.config = config.C
	handler.api = cpapi.NewCpApi()
	command := &cobra.Command{
		Use:     "tidestart",
		Aliases: []string{"st", "start"},
		Short:   "Starts a new tide",
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			validateConfig()

			benchmrk := benchmark.NewCmdBenchmark()
			benchmrk.Start("tidestart")

			checkErr(handler.Handle())
			_, err := benchmrk.StopAndLog()
			checkErr(err)
		},
		Example: startTideExample,
	}
	return command
}

type TideStartHandle struct {
	command *cobra.Command
	config  config.ConfigProvider
	stdout  io.Writer
	api     cpapi.CpApiProvider
}

func (h *TideStartHandle) Handle() error {
	return nil
}
