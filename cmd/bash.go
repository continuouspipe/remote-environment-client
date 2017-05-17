package cmd

import (
	"fmt"

	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/spf13/cobra"
)

var bashInteractiveFullExample = fmt.Sprintf(`%s bash --interactive ([-i]) --environment ([-e]) php-example-cpdev-foo --service ([-s]) web --flow-id ([-f]) 1268cc54-b265-11e6-b835-0c360641bb54`, config.AppName)

var bashExample = fmt.Sprintf(`
Default Mode:

	# bash into the default remote container
	%[1]s bash
	%[1]s ba

	# bash into a different environment and service within the same flow
	%[1]s bash -e techup-dev-user -s web

Interactive Mode:

	# bash into a different environment (without knowing which one yet)
	%[1]s bash --interactive
	%[1]s bash -i

	# bash into a different environment and service in a different flow
	%[2]s`, config.AppName, bashInteractiveFullExample)

//BashCmdName is the name identifier for the exec command
const BashCmdName = "bash"

func NewBashCmd() *cobra.Command {
	bashcmd := NewExecCmd()
	bashcmd.Use = BashCmdName
	bashcmd.Aliases = []string{"ba"}
	bashcmd.Short = "Open a bash session in the remote environment container"
	bashcmd.Long = `This will remotely connect to a bash session onto the default container specified
during setup but you can specify another container to connect to. `
	bashcmd.Example = bashExample
	execRun := bashcmd.Run
	bashcmd.Run = func(cmd *cobra.Command, args []string) {
		execRun(cmd, []string{"/bin/bash"})
	}

	return bashcmd
}
