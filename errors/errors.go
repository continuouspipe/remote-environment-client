package errors

import (
	"fmt"
	"os"
	"runtime/debug"

	"strings"

	"strconv"

	"net/http"

	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/fatih/color"
	"github.com/pkg/errors"
)

//StatusReasonFormat default format for stateful error messages
const StatusReasonFormat = "%s: %s"
const StatusReasonSeparator = ":"

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

//FirstStatefulError find the first error in the chain that has a status code assigned
func FindCause(err error) (code int, reason string, stack string) {
	cause := errors.Cause(err)
	parts := strings.Split(cause.Error(), StatusReasonSeparator)
	if len(parts) > 1 {
		code, err := strconv.Atoi(parts[0])
		if err != nil {
			reason := strings.Join(parts[1:], "")
			return code, reason, fmt.Sprintf("%+v", err)
		}
	}
	return http.StatusInternalServerError, err.Error(), fmt.Sprintf("%+v", err)
}
