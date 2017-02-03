package osapi

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/continuouspipe/remote-environment-client/cplogs"
)

//reduced version of the os.Command to pass in minimal information required by the osapi package
type SCommand struct {
	Name   string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

//Executes a command and waits for it to finish
func CommandExec(scmd SCommand, arg ...string) (string, error) {
	cmd := exec.Command(scmd.Name, arg...)
	cmd.Stdin = scmd.Stdin
	cmd.Stderr = scmd.Stderr

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

//execute the command and write output into log file
func CommandExecL(scmd SCommand, arg ...string) error {
	cmd := exec.Command(scmd.Name, arg...)
	cmd.Stdin = scmd.Stdin
	cmd.Stderr = scmd.Stderr

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
		if cmd.Stdout != nil {
			fmt.Fprintln(cmd.Stdout, line)
		}
	}
	return nil
}

//Start a process and waits for it to finish
//it redirects the cmd Stdin Stdout and Stderr in the current process ones
func StartProcess(scmd SCommand, killProcess chan bool, arg ...string) error {
	cmd := exec.Command(scmd.Name, arg...)
	cmd.Stderr = scmd.Stderr
	cmd.Stdin = scmd.Stdin
	cmd.Stdout = scmd.Stdout

	cplogs.V(5).Infof("executing command path: %s with arguments: %#v", cmd.Path, cmd.Args)
	cplogs.Flush()
	err := cmd.Start()
	if err != nil {
		return err
	}

	cplogs.V(5).Infoln("wait for command to finish")

	if killProcess == nil {
		killProcess = make(chan bool, 1)
	}

	errChan := make(chan error, 1)

	go func() {
		errChan <- cmd.Wait()
	}()

	select {
	case <-killProcess:
		if err := cmd.Process.Kill(); err != nil {
			cplogs.V(3).Infof("failed to kill: %s", err.Error())
		}
		return err
	case err := <-errChan:
		if err != nil {
			cplogs.V(3).Infof("command error: %s", err.Error())
		}
	}

	return nil
}
