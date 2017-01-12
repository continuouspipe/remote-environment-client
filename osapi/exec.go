package osapi

import (
	"os/exec"
	"syscall"
	"os"
)

func CommandExec(name string, arg ...string) string {
	cmd := exec.Command(name, arg...)
	return executeCmd(cmd)
}

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
		panic(err)
	}
	return string(out[:])
}
