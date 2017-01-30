package exec

import (
	"github.com/continuouspipe/remote-environment-client/config"
	"github.com/continuouspipe/remote-environment-client/osapi"
)

type Spawner interface {
	CommandExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) (string, error)
}

type Executor interface {
	StartProcess(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) error
}

// the local type executes and spawn commands locally
type Local struct{}

func NewLocal() *Local {
	return &Local{}
}

// executes a command (spawn) on a specific pod
func (l Local) CommandExec(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) (string, error) {
	args := l.getAllArgs(kubeConfigKey, environment, pod, execCmdArgs...)
	return osapi.CommandExec(config.AppName, args...)
}

// executes a system call (exec) on a specific pod
func (l Local) StartProcess(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) error {
	args := l.getAllArgs(kubeConfigKey, environment, pod, execCmdArgs...)
	return osapi.StartProcess(config.AppName, args...)
}

// sets all the flags required to execute a command inside a container
func (l Local) getAllArgs(kubeConfigKey string, environment string, pod string, execCmdArgs ...string) []string {
	kubeCmdArgs := []string{
		config.KubeCtlName,
		"--context=" + kubeConfigKey,
		"--namespace=" + environment,
		"exec",
		"-it",
		pod,
		"--",
	}

	allArgs := append(kubeCmdArgs, execCmdArgs...)
	return allArgs
}
