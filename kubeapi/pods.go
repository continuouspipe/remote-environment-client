package kubeapi

import (
	"errors"
	"strings"

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
	config, err := ReadConfig(context)

	if err != nil {
		return nil, err
	}

	client, err := CreateClient(config)

	if err != nil {
		return nil, err
	}

	return client.Core().Pods(environment).List(v1.ListOptions{})
}

func FindPodByContainer(context string, environment string, container string) (v1.Pod, error) {
	pods, err := FetchPods(context, environment)

	if err != nil {
		return v1.Pod{}, err
	}

	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.GetName(), container) {
			return pod, nil
		}
	}

	return v1.Pod{}, errors.New("Pods were found but not for the container specified")

}
