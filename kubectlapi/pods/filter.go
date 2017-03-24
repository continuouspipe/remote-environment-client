package pods

import (
	"fmt"
	"k8s.io/kubernetes/pkg/api"
	"strings"
)

type Filter interface {
	ByService(*api.PodList, string) (*api.Pod, error)
}

type KubePodsFilter struct{}

func NewKubePodsFilter() *KubePodsFilter {
	return &KubePodsFilter{}
}

func (p KubePodsFilter) ByService(pods *api.PodList, service string) (*api.Pod, error) {
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.GetName(), service) {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf(fmt.Sprintf("Pods were found but not for the service name (%s) specified", service))
}
