package kubectlapi

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func readConfig(context string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: context}).ClientConfig()
}

func createClient(config *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(config)
}

func FetchPods(context string, environment string) (*v1.PodList, error) {
	config, err := readConfig(context)

	if err != nil {
		return nil, err
	}

	client, err := createClient(config)

	if err != nil {
		return nil, err
	}

	pods, err := client.Core().Pods(environment).List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pods, nil
}
