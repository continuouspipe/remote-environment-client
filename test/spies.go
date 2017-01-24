package test

import (
	"fmt"
	"github.com/continuouspipe/remote-environment-client/config"
)

//A map that stores a list of function arguments [argumentName] => value (any type)
type Arguments map[string]interface{}

//Function is a struct where you can set the name and add a slice Arguments ([]Argument) for each call
type Function struct {
	Name      string
	Arguments Arguments
}

//Generic struct that can be embedded by any struct that wants to keep track to what function was called and with which args
type Spy struct {
	calledFunctions []Function
	commandExec     func() string
}

//Returns the Function call element for the given functionName, this is useful when a type has received multiple functions
func (spy *Spy) FirstCallsFor(functionName string) *Function {
	for _, call := range spy.calledFunctions {
		if call.Name == functionName {
			return &call
		}
	}
	return nil
}

//returns how many times the given function has been called
func (spy *Spy) CallsCountFor(functionName string) int {
	count := 0
	for _, call := range spy.calledFunctions {
		if call.Name != functionName {
			continue
		}
		count++
	}
	return count
}

//Spy to mock the LocalExecutor
type SpyLocalExecutor struct {
	Spy
}

func GetSpyLocalExecutor() *SpyLocalExecutor {
	return &SpyLocalExecutor{}
}

func (m *SpyLocalExecutor) SysCallExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) {
	args := make(Arguments)
	args["kubeConfigKey"] = kubeConfigKey
	args["environment"] = environment
	args["pod"] = pod
	args["execCmdArgs"] = execCmdArgs

	function := &Function{Name: "SysCallExec", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)
}

func (m *SpyLocalExecutor) CommandExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) string {
	args := make(Arguments)
	args["kubeConfigKey"] = kubeConfigKey
	args["environment"] = environment
	args["pod"] = pod
	args["execCmdArgs"] = execCmdArgs

	function := &Function{Name: "CommandExec", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.commandExec()
}

func (m *SpyLocalExecutor) SpyCommandExec(mocked func() string) {
	m.commandExec = mocked
}

//Spy for RsyncFetch
type SpyRsyncFetch struct {
	Spy
}

func GetSpyRsyncFetch() *SpyRsyncFetch {
	return &SpyRsyncFetch{}
}

func (r *SpyRsyncFetch) Fetch(kubeConfigKey string, environment string, pod string) bool {
	args := make(Arguments)
	args["kubeConfigKey"] = kubeConfigKey
	args["environment"] = environment
	args["pod"] = pod
	function := &Function{Name: "Fetch", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
	return true
}

//Spy for YamlWriter
type SpyYamlWriter struct {
	Spy
}

func GetSpyYamlWriter() *SpyYamlWriter {
	return &SpyYamlWriter{}
}

func (m *SpyYamlWriter) Save(settings *config.ApplicationSettings) bool {

	copySettings := *settings

	args := make(Arguments)
	args["settings"] = &copySettings

	fmt.Println(args["settings"])

	function := &Function{Name: "Save", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)
	return true
}