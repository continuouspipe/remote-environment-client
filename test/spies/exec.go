package spies

import (
	kexec "github.com/continuouspipe/remote-environment-client/kubectlapi/exec"
)

//Spy to mock the LocalExecutor
type SpyLocalExecutor struct {
	Spy
	startProcess func() error
	commandExec  func() (string, error)
}

func NewSpyLocalExecutor() *SpyLocalExecutor {
	return &SpyLocalExecutor{}
}

func (m *SpyLocalExecutor) StartProcess(kscmd kexec.KSCommand, execCmdArgs ...string) error {
	args := make(Arguments)
	args["kscmd"] = kscmd
	args["execCmdArgs"] = execCmdArgs

	function := &Function{Name: "StartProcess", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.startProcess()
}

func (m *SpyLocalExecutor) CommandExec(kscmd kexec.KSCommand, execCmdArgs ...string) (string, error) {
	args := make(Arguments)
	args["kscmd"] = kscmd
	args["execCmdArgs"] = execCmdArgs

	function := &Function{Name: "CommandExec", Arguments: args}
	m.calledFunctions = append(m.calledFunctions, *function)

	return m.commandExec()
}

func (m *SpyLocalExecutor) MockCommandExec(mocked func() (string, error)) {
	m.commandExec = mocked
}

func (m *SpyLocalExecutor) MockStartProcess(mocked func() error) {
	m.startProcess = mocked
}
