package cmd

import (
	"os"

	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	kubectlcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func init() {
	kubeCtlCommand := kubectlcmd.NewKubectlCommand(kubectlcmdutil.NewFactory(nil), os.Stdin, os.Stdout, os.Stderr)
	kubeCtlCommand.Hidden = true
	RootCmd.AddCommand(kubeCtlCommand)
}
