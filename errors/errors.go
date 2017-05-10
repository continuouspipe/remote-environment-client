package errors

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/fatih/color"
)

const StatusReasonFormat = "%s: %s"

func CheckErr(err error) {
	if err == nil {
		return
	}
	ExitWithMessage(err.Error())
}

func ExitWithMessage(message string) {
	color.Set(color.FgRed)
	fmt.Println(message)

	stack := debug.Stack()
	cplogs.V(4).Info(string(stack[:]))
	color.Unset()
	cplogs.Flush()
	os.Exit(1)
}
