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

func CheckErrorList(errList ErrorList) {
	if errList.HasErrors() {
		ExitWithMessage(errList.String())
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

type ErrorList struct {
	Items []error
}

func (el *ErrorList) Add(err error) {
	el.Items = append(el.Items, err)
}

func (el ErrorList) HasErrors() bool {
	return len(el.Items) > 0
}

func (el ErrorList) String() (err string) {

	if len(el.Items) <= 0 {
		return
	}

	err = fmt.Sprintf("An error occured: %s", el.Items[0].Error())
	if len(el.Items) == 1 {
		return
	}

	err = err + fmt.Sprintf("\nError stack:\n")

	for key, error := range el.Items {
		err = err + fmt.Sprintf("\n[%d] %s", key, error.Error())
	}
	return
}
