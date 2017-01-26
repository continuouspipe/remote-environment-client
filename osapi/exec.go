package osapi

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/continuouspipe/remote-environment-client/cplogs"
	"bufio"
	"fmt"
)

//Executes a command and waits for it to finish
func CommandExec(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	return executeCmd(cmd)
}

//execute the command and write output into log file
//and optionally writes it on a file descriptor (e.g. os.Stdout)
func CommandExecL(name string, file *os.File, arg ...string) error {
	cmd := exec.Command(name, arg...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cplogs.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		cplogs.Fatal(err)
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		cplogs.V(5).Infoln(line)
		if file != nil {
			fmt.Fprintln(file, line)
		}
	}
	return nil
}

//Exec a command and then continues without waiting
func SysCallExec(name string, arg ...string) {
	appBinPath, lookErr := exec.LookPath(name)
	if lookErr != nil {
		cplogs.V(5).Infof("path to binary file %s not found", name)
		panic(lookErr)
	}

	env := os.Environ()

	//syscall.Exec requires the first argument to be the app-name
	allArgs := append([]string{name}, arg...)

	cplogs.V(5).Infof("executing command path: %#v with arguments: %#v", appBinPath, allArgs)
	cplogs.Flush()
	execErr := syscall.Exec(appBinPath, allArgs, env)
	if execErr != nil {
		cplogs.V(3).Infof("command error: %#v", execErr.Error())
		cplogs.Flush()
		panic(execErr)
	}
}

func executeCmd(cmd *exec.Cmd) (string, error) {
	cplogs.V(5).Infof("executing command path: %#v with arguments: %#v", cmd.Path, cmd.Args)
	cplogs.Flush()
	out, err := cmd.Output()
	if err != nil {
		cplogs.V(3).Infof("command error: %#v", err.Error())
		cplogs.Flush()
		return "", err
	}
	sout := string(out[:])
	cplogs.V(7).Infof("command output as string: %s", sout)
	cplogs.Flush()
	//remove newline and space from string
	return strings.Trim(sout, "\n "), nil
}
