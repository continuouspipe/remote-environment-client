package osapi

import (
	"os/exec"
	"syscall"
	"os"
)

//Executes a command and waits for it to finish
func CommandExec(name string, arg ...string) string {
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

func executeCmd(cmd *exec.Cmd) string {
	out, err := cmd.Output()
	if err != nil {
		panic(err.Error())
	}
	return string(out[:])
}
