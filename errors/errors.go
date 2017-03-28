package errors

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/fatih/color"
	"os"
	"runtime/debug"
)

func CheckErr(err error) {
	if err == nil {
		return
	}

	switch t := err.(type) {
	case *ErrorList:
		if t == nil {
			return
		}
		if len(t.Items()) == 0 {
			return
		}
		ExitWithMessage(err.Error())
	case ErrorList:
		if len(t.Items()) == 0 {
			return
		}
		ExitWithMessage(err.Error())
	default:
		ExitWithMessage(err.Error())
	}

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

type ErrorListProvider interface {
	error
	Add(elems ...error)
	Items() []error
}

type ErrorList struct {
	errors []error
}

func NewErrorList() *ErrorList {
	return &ErrorList{}
}

func (el *ErrorList) Add(elems ...error) {
	el.errors = append(el.errors, elems...)
}

func (el *ErrorList) AddErrorf(format string, a ...interface{}) {
	err := fmt.Errorf(format, a...)
	el.errors = append(el.errors, err)
}

func (el ErrorList) Items() []error {
	return el.errors
}

func (el ErrorList) Error() (err string) {
	if len(el.errors) <= 0 {
		return
	}

	err = fmt.Sprintf("An error occured: %s", el.errors[0].Error())
	if len(el.errors) == 1 {
		return
	}

	err = err + fmt.Sprintf("\nError stack:\n")

	for key, item := range el.errors {
		if key == 0 {
			continue
		}
		err = err + fmt.Sprintf("\n[%d] %s", key, item.Error())
	}
	return
}
