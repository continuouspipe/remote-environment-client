package cmd

import (
	"github.com/spf13/cobra"
	"github.com/continuouspipe/remote-environment-client/config"
	"fmt"
	"encoding/base64"
	"strings"
	"github.com/continuouspipe/kube-proxy/cplogs"
)

func NewInitCmd() *cobra.Command {
	//settings := config.C
	handler := &InitHandler{}
	command := &cobra.Command{
		Use:     "init [cp-remote-token]",
		Aliases: []string{"in"},
		Short:   "Initialises the remote environment",
		Long:    ``,
		Run: func(cmd *cobra.Command, args []string) {
			checkErr(handler.Complete(args))
			checkErr(handler.Validate())
			checkErr(handler.Handle())
		},
		Example: portforwardExample,
	}
	return command
}

type InitHandler struct {
	command *cobra.Command
	config  *config.Config
	token   string
}

// Complete verifies command line arguments and loads data from the command environment
func (i *InitHandler) Complete(argsIn []string) error {

	fmt.Println(len(argsIn))
	if len(argsIn) > 0 && argsIn[0] != "" {
		fmt.Println(argsIn[0])
		i.token = argsIn[0]
		return nil
	}
	return fmt.Errorf("Invalid token. Please go to continouspipe.io to obtain a valid token.")
}

// Validate checks that the token provided has at least 4 values comma separated
func (i InitHandler) Validate() error {
	decodedToken, err := base64.StdEncoding.DecodeString(i.token)
	if err != nil {
		return fmt.Errorf("Malformed token. Please go to continouspipe.io to obtain a valid token.")
	}
	splitToken := strings.Split(string(decodedToken), ",")
	if len(splitToken) != 4 {
		cplogs.V(4).Infof("Token provided %s has %d parts, expected 4", splitToken, len(splitToken))
		return fmt.Errorf("Malformed token. Please go to continouspipe.io to obtain a valid token.")
	}
	return nil
}

//Handle Executes the initialization
func (i *InitHandler) Handle() error {
	return nil
}
