package cmd

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/spf13/cobra"
	"runtime"
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
	fmt.Printf("Current version: %s (%s-%s)\n", config.CurrentVersion, runtime.GOOS, runtime.GOARCH)
}
