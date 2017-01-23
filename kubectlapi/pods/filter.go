package pods

import (
	"fmt"
	"strings"

	"k8s.io/client-go/pkg/api/v1"
)

type Filter interface {
	ByService(*v1.PodList, string) (*v1.Pod, error)
}

type KubePodsFilter struct{}

func NewKubePodsFilter() *KubePodsFilter {
	return &KubePodsFilter{}
}

func (p KubePodsFilter) ByService(pods *v1.PodList, service string) (*v1.Pod, error) {
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.GetName(), service) {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf(fmt.Sprintf("Pods were found but not for the service name (%s) specified", service))
}
