package test

import "github.com/continuouspipe/remote-environment-client/config"

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
	commandExec     func() (string, error)
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

func NewSpyLocalExecutor() *SpyLocalExecutor {
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

func (m *SpyLocalExecutor) CommandExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) (string, error) {
	args := make(Arguments)
	args["kubeConfigKey"] = kubeConfigKey
	args["environment"] = environment
	args["pod"] = pod
	args["execCmdArgs"] = execCmdArgs

	function := &Function{Name: "CommandExec", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.commandExec()
}

func (m *SpyLocalExecutor) SpyCommandExec(mocked func() (string, error)) {
	m.commandExec = mocked
}

//Spy for RsyncFetch
type SpyRsyncFetch struct {
	Spy
}

func NewSpyRsyncFetch() *SpyRsyncFetch {
	return &SpyRsyncFetch{}
}

func (r *SpyRsyncFetch) Fetch(kubeConfigKey string, environment string, pod string) error {
	args := make(Arguments)
	args["kubeConfigKey"] = kubeConfigKey
	args["environment"] = environment
	args["pod"] = pod
	function := &Function{Name: "Fetch", Arguments: args}
	r.calledFunctions = append(r.calledFunctions, *function)
	return nil
}

//Spy for YamlWriter
type SpyYamlWriter struct {
	Spy
}

func NewSpyYamlWriter() *SpyYamlWriter {
	return &SpyYamlWriter{}
}

func (m *SpyYamlWriter) Save(settings *config.ApplicationSettings) bool {
	copySettings := *settings

	args := make(Arguments)
	args["settings"] = &copySettings

	function := &Function{Name: "Save", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)
	return true
}

//Spy for Commit
type SpyCommit struct {
	Spy
}

func NewSpyCommit() *SpyCommit {
	return &SpyCommit{}
}

func (s *SpyCommit) Commit(message string) (string, error) {
	args := make(Arguments)
	args["message"] = message
	function := &Function{Name: "Commit", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return "", nil
}

//Spy for Push
type SpyPush struct {
	Spy
}

func NewSpyPush() *SpyPush {
	return &SpyPush{}
}

func (s *SpyPush) Push(localBranch string, remoteName string, remoteBranch string) (string, error) {
	args := make(Arguments)
	args["localBranch"] = localBranch
	args["remoteName"] = remoteName
	args["remoteBranch"] = remoteBranch

	function := &Function{Name: "Push", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return "", nil
}

func (s *SpyPush) DeleteRemote(remoteName string, remoteBranch string) (string, error) {
	args := make(Arguments)
	args["remoteName"] = remoteName
	args["remoteBranch"] = remoteBranch

	function := &Function{Name: "DeleteRemote", Arguments: args}
	s.calledFunctions = append(s.calledFunctions, *function)
	return "", nil
}
