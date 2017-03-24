package pods

import (
	"github.com/continuouspipe/remote-environment-client/kubectlapi"
	"k8s.io/kubernetes/pkg/api"
)

type Finder interface {
	FindAll(user string, apiKey string, address string, environment string) (*api.PodList, error)
}

type KubePodsFind struct{}

func NewKubePodsFind() *KubePodsFind {
	return &KubePodsFind{}
}

func (p KubePodsFind) FindAll(user string, apiKey string, address string, environment string) (*api.PodList, error) {
	config, err := kubectlapi.GetNonInteractiveDeferredLoadingClientConfig(user, apiKey, address).ClientConfig()

	if err != nil {
		return nil, err
	}

	client, err := kubectlapi.CreateClient(config)

	if err != nil {
		return nil, err
	}

	return client.Core().Pods(environment).List(api.ListOptions{})
}
