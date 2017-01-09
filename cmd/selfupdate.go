package cmd

import (
	"fmt"

	envconfig "github.com/continuouspipe/remote-environment-client/config"
	"github.com/spf13/cobra"
	"github.com/sanbornm/go-selfupdate/selfupdate"
)

var selfUpdateCmd = &cobra.Command{
	Use:   "selfupdate",
	Short: "Download and Upgrade to the latest version",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		handler := &SelfUpdateHandle{cmd}
		handler.Handle(args)
	},
}

type SelfUpdateHandle struct {
	Command *cobra.Command
}

func (h *SelfUpdateHandle) Handle(args []string) {
	var updater = &selfupdate.Updater{
		// Manually update the const, or set it using `go build -ldflags="-X main.VERSION=<newver>" -o cp-remote remote-environment-client/main.go`
		CurrentVersion: envconfig.CurrentVersion,
		// The server hosting `$CmdName/$GOOS-$ARCH.json` which contains the checksum for the binary
		ApiURL:         "https://raw.githubusercontent.com/continuouspipe/remote-environment-client/gh-pages/downloads/",
		// The server hosting the zip file containing the binary application which is a fallback for the patch method
		BinURL:         "https://raw.githubusercontent.com/continuouspipe/remote-environment-client/gh-pages/downloads/",
		// The server hosting the binary patch diff for incremental updates
		DiffURL:        "https://raw.githubusercontent.com/continuouspipe/remote-environment-client/gh-pages/downloads/",
		// The directory created by the app when run which stores the cktime file
		Dir:            "update/",
		// The app name which is appended to the ApiURL to look for an update
		CmdName:        "cp-remote-latest",
	}

	fmt.Printf("Currently version %v\n", updater.CurrentVersion)
	err := updater.BackgroundRun()
	checkErr(err)
	fmt.Printf("Next run version %v\n", updater.Info.Version)
}

func init() {
	RootCmd.AddCommand(selfUpdateCmd)
}
