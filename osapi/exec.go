package osapi

import (
	"os"
	"os/exec"
	"strings"

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
func StartProcess(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		return err
	}
	cplogs.V(5).Infoln("wait for command to finish")
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
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
