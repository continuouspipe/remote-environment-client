package error

import (
	"fmt"
	"runtime/debug"
	"os"
	"github.com/fatih/color"
	"github.com/continuouspipe/remote-environment-client/cplogs"
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
