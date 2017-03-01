package error

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/fatih/color"
	"os"
	"runtime/debug"
)

func CheckErr(err error) {
	if err != nil {
		ExitWithMessage(err.Error())
	}
}

func ExitWithMessage(message string) {
	color.Set(color.FgRed)
	fmt.Println("ERROR: " + message)

	stack := debug.Stack()
	cplogs.V(4).Info(string(stack[:]))
	color.Unset()
	cplogs.Flush()
	os.Exit(1)
}
