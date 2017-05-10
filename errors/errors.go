package errors

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/fatih/color"
)

//StatusReasonFormat default format for stateful error messages
const StatusReasonFormat = "%s: %s"

//CheckErr calls ExitWithMessage when there is an error
func CheckErr(err error) {
	if err == nil {
		return
	}
	ExitWithMessage(err.Error())
}

//ExitWithMessage print and write the stacktrace on the logs
func ExitWithMessage(message string) {
	color.Set(color.FgRed)
	fmt.Println(message)

	stack := debug.Stack()
	cplogs.V(4).Info(string(stack[:]))
	color.Unset()
	cplogs.Flush()
	os.Exit(1)
}
