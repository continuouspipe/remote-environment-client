package cmd

import (
	"fmt"
	"runtime"

	envconfig "github.com/continuouspipe/remote-environment-client/config"
	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"ve"},
		Short:   "Show current version number",
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			handler := &VersionHandle{cmd}
			handler.Handle(args)
		},
	}
}

type VersionHandle struct {
	Command *cobra.Command
}

func (h *VersionHandle) Handle(args []string) {
	fmt.Printf("Current version: %s (%s-%s)\n", envconfig.CurrentVersion, runtime.GOOS, runtime.GOARCH)
}
