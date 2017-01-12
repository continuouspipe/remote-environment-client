package kubectlapi

import (
	"os/exec"
)

const appName = "cp-remote-go"
const kubeCtlName = "kubectl"

func executeCmd(cmd *exec.Cmd) string {
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	return string(out[:])
}
