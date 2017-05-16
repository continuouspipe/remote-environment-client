package errors

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"

	"regexp"

	"github.com/continuouspipe/remote-environment-client/cplogs"
	"github.com/fatih/color"
)

//StatefulErrorMessage groups an error code with a message
type StatefulErrorMessage struct {
	Code      int
	Message   string
	Format    string `default:"%d: %s"`
	Separator string `default:": "`
}

//NewStatefulErrorMessage return a StatefulErrorMessage
func NewStatefulErrorMessage(code int, msg string) *StatefulErrorMessage {
	sem := &StatefulErrorMessage{Code: code, Message: msg}
	sem.appyDefaultsFromTags()
	return sem
}

//NewStatefulErrorMessageFromString create a StatefulErrorMessage from string
func NewStatefulErrorMessageFromString(str string) *StatefulErrorMessage {
	sem := &StatefulErrorMessage{}
	sem.appyDefaultsFromTags()
	parts := strings.Split(str, sem.Separator)
	if len(parts) > 1 {
		code, err := strconv.Atoi(parts[0])
		if err != nil {
			cplogs.V(5).Infof("error code convertion failed, error %s", err)
			cplogs.Flush()
			return nil
		}
		sem.Code = code
		sem.Message = strings.Join(parts[1:], "")
		return sem
	}
	return nil
}

//String converts a stateful error message to string
func (s StatefulErrorMessage) String() string {
	return fmt.Sprintf(s.Format, s.Code, s.Message)
}

func (s *StatefulErrorMessage) appyDefaultsFromTags() {
	field, ok := reflect.TypeOf(s).Elem().FieldByName("Format")
	if ok {
		s.Format = field.Tag.Get("default")
	}
	field, ok = reflect.TypeOf(s).Elem().FieldByName("Separator")
	if ok {
		s.Separator = field.Tag.Get("default")
	}
}

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

//FindCause finds the first error in the chain that has a status code assigned
func FindCause(err error) (code int, reason string, stack string) {
	cause := StatefulCause(err)

	if cause != nil {
		sem := NewStatefulErrorMessageFromString(cause.Error())
		if sem != nil {
			return sem.Code, sem.Message, fmt.Sprintf("%+v", err)
		}
	}

	return http.StatusInternalServerError, err.Error(), fmt.Sprintf("%+v", err)
}

// StatefulCause attempts to find the cause of the error that has a state associated, if it finds it returns it, otherwise returns the cause itself
func StatefulCause(err error) error {
	type causer interface {
		Cause() error
	}

	var lastCause error
	var statefulCause error

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()

		r, compErr := regexp.Compile(`\d+:`)
		if compErr != nil {
			break
		}
		if r.MatchString(err.Error()) {
			statefulCause = err
		} else {
			lastCause = err
		}
	}
	if statefulCause != nil {
		return statefulCause
	}
	if lastCause != nil {
		return lastCause
	}
	return err
}
