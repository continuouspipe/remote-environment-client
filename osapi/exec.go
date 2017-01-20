package osapi

import (
	"os"
	"strings"
	"syscall"
	"os/exec"

	"github.com/continuouspipe/remote-environment-client/cplogs"
)

//Executes a command and waits for it to finish
func CommandExec(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	return executeCmd(cmd)
}

//Exec a command and then continues without waiting
func SysCallExec(name string, arg ...string) {
	appBinPath, lookErr := exec.LookPath(name)
	if lookErr != nil {
		panic(lookErr)
	}

	env := os.Environ()

	//syscall.Exec requires the first argument to be the app-name
	allArgs := append([]string{name}, arg...)

	execErr := syscall.Exec(appBinPath, allArgs, env)
	if execErr != nil {
		panic(execErr)
	}
}

func executeCmd(cmd *exec.Cmd) (string, error) {
	cplogs.V(5).Infof("executing command path: %#v with arguments: %#v", cmd.Path, cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		cplogs.V(3).Infof("command error: %#v", err)
		return "", err
	}
	sout := string(out[:])
	cplogs.V(5).Infof("command output as string: %s", sout)
	//remove newline and space from string
	return strings.Trim(sout, "\n "), nil
}
