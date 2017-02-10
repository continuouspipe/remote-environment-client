package exec

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/osapi"
	"io"
)

type KSCommand struct {
	KubeConfigKey string
	Environment   string
	Pod           string
	Stdin         io.Reader
	Stdout        io.Writer
	Stderr        io.Writer
}

type Spawner interface {
	CommandExec(kscmd KSCommand, execCmdArgs ...string) (string, error)
}

type Executor interface {
	StartProcess(kscmd KSCommand, execCmdArgs ...string) error
}

// the local type executes and spawn commands locally
type Local struct{}

func NewLocal() *Local {
	return &Local{}
}

// executes a command (spawn) on a specific pod
func (l Local) CommandExec(kscmd KSCommand, execCmdArgs ...string) (string, error) {
	args := l.getAllArgs(kscmd, execCmdArgs...)

	return osapi.CommandExec(getAppScmd(kscmd), args...)
}

// executes a system call (exec) on a specific pod
func (l Local) StartProcess(kscmd KSCommand, execCmdArgs ...string) error {
	args := l.getAllArgs(kscmd, execCmdArgs...)

	return osapi.StartProcess(getAppScmd(kscmd), nil, args...)
}

// sets all the flags required to execute a command inside a container
func (l Local) getAllArgs(kscmd KSCommand, execCmdArgs ...string) []string {
	kubeCmdArgs := []string{
		config.KubeCtlName,
		"--context=" + kscmd.KubeConfigKey,
		"--namespace=" + kscmd.Environment,
		"exec",
		"-it",
		kscmd.Pod,
		"--",
	}

	allArgs := append(kubeCmdArgs, execCmdArgs...)
	return allArgs
}

func getAppScmd(kscmd KSCommand) osapi.SCommand {
	scmd := osapi.SCommand{}
	scmd.Name = config.AppName
	scmd.Stdin = kscmd.Stdin
	scmd.Stdout = kscmd.Stdout
	scmd.Stderr = kscmd.Stderr
	return scmd
}
