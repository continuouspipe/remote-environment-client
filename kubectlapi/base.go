package kubectlapi

import (
	"os/exec"
)

func executeCmd(cmd *exec.Cmd) string {
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	return string(out[:])
}
