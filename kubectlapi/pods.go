package kubectlapi

import (
	"errors"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"fmt"
)

func readConfig(kubeConfigKey string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: kubeConfigKey}).ClientConfig()
}

func createClient(config *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(config)
}

func FetchPods(kubeConfigKey string, environment string) (*v1.PodList, error) {
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

func FindPodByService(kubeConfigKey string, environment string, service string) (*v1.Pod, error) {
	pods, err := FetchPods(kubeConfigKey, environment)

	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.GetName(), service) {
			return &pod, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("Pods were found but not for the service name (%s) specified", service))

}
