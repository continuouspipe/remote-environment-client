package pods

import (
	"k8s.io/kubernetes/pkg/api"
	"strings"
)

type Filter interface {
	List(pods api.PodList) KubePodsFilter
	ByService(service string) KubePodsFilter
	ByStatus(status string) KubePodsFilter
	First() *api.Pod
}

type KubePodsFilter struct {
	podList api.PodList
}

func NewKubePodsFilter() *KubePodsFilter {
	return &KubePodsFilter{}
}

func (p *KubePodsFilter) List(podList api.PodList) KubePodsFilter {
	p.podList = podList
	return *p
}

func (p KubePodsFilter) First() *api.Pod {
	if len(p.podList.Items) > 0 {
		return &p.podList.Items[0]
	}
	return nil
}

func (p KubePodsFilter) ByService(service string) KubePodsFilter {
	filteredPodItems := p.podList.Items[:0]
	for _, pod := range p.podList.Items {
		if strings.HasPrefix(pod.GetName(), service) {
			filteredPodItems = append(filteredPodItems, pod)
		}
	}
	p.podList.Items = filteredPodItems
	return p
}

func (p KubePodsFilter) ByStatus(status string) KubePodsFilter {
	filteredPodItems := p.podList.Items[:0]
	for _, pod := range p.podList.Items {
		if pod.Status.Phase == statusToPodPhase(status) {
			filteredPodItems = append(filteredPodItems, pod)
		}
	}
	p.podList.Items = filteredPodItems
	return p
}

func statusToPodPhase(status string) api.PodPhase {
	switch status {
	case "Pending":
		return api.PodPending
	case "Running":
		return api.PodRunning
	case "Succeeded":
		return api.PodSucceeded
	case "Failed":
		return api.PodFailed
	case "Unknown":
		return api.PodUnknown
	}
	return ""
}
