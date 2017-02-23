package services

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ServiceFinder interface {
	FindAll(kubeConfigKey string, environment string) (*v1.ServiceList, error)
}

type KubeService struct{}

func NewKubeService() *KubeService {
	return &KubeService{}
}

func (p KubeService) FindAll(kubeConfigKey string, environment string) (*v1.ServiceList, error) {
	config, err := readConfig(kubeConfigKey)

	if err != nil {
		return nil, err
	}

	client, err := createClient(config)

	if err != nil {
		return nil, err
	}

	return client.Core().Services(environment).List(v1.ListOptions{})
}

func readConfig(kubeConfigKey string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: kubeConfigKey}).ClientConfig()
}

func createClient(config *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(config)
}
