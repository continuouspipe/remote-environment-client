package pods

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Finder interface {
	FindAll(kubeConfigKey string, environment string) (*v1.PodList, error)
}

type KubePodsFind struct{}

func NewKubePodsFind() *KubePodsFind {
	return &KubePodsFind{}
}

func (p KubePodsFind) FindAll(kubeConfigKey string, environment string) (*v1.PodList, error) {
	config, err := readConfig(kubeConfigKey)

	if err != nil {
		return nil, err
	}

	client, err := createClient(config)

	if err != nil {
		return nil, err
	}

	return client.Core().Pods(environment).List(v1.ListOptions{})
}

func readConfig(kubeConfigKey string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: kubeConfigKey}).ClientConfig()
}

func createClient(config *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(config)
}
