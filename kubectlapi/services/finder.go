package services

import (
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"k8s.io/kubernetes/pkg/api"
)

type ServiceFinder interface {
	FindAll(user string, apiKey string, address string, environment string) (*api.ServiceList, error)
}

type KubeService struct{}

func NewKubeService() *KubeService {
	return &KubeService{}
}

func (p KubeService) FindAll(user string, apiKey string, address string, environment string) (*api.ServiceList, error) {
	config, err := kubectlapi.GetNonInteractiveDeferredLoadingClientConfig(user, apiKey, address).ClientConfig()

	if err != nil {
		return nil, err
	}

	client, err := kubectlapi.CreateClient(config)

	if err != nil {
		return nil, err
	}

	return client.Core().Services(environment).List(api.ListOptions{})
}
