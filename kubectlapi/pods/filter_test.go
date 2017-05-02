package pods

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/util/node"
)

func TestKubePodsFilter(t *testing.T) {
	podPending := api.Pod{}
	podPending.Name = "mysql-881010915-wpndh"
	podPending.Status = api.PodStatus{Phase: api.PodPending}

	podFailed := api.Pod{}
	podFailed.Name = "web-151435215-nsdaf"
	podFailed.Status = api.PodStatus{Phase: api.PodFailed}

	podUnknown := api.Pod{}
	podUnknown.Name = "web-981327404-bargs"
	podUnknown.Status = api.PodStatus{Phase: api.PodUnknown}

	podTerminating := api.Pod{}
	podTerminating.Name = "web-9263482738-axiwjs"
	podTerminating.Status = api.PodStatus{Phase: api.PodRunning}
	podTerminating.Status = api.PodStatus{Phase: api.PodRunning, Reason: node.NodeUnreachablePodReason}
	podTerminating.DeletionTimestamp = &unversioned.Time{}

	podRunning := api.Pod{}
	podRunning.Name = "web-812374193-mxiwy"
	podRunning.Status = api.PodStatus{Phase: api.PodRunning}

	podSucceeded := api.Pod{}
	podSucceeded.Name = "web-989823427-cosjd"
	podSucceeded.Status = api.PodStatus{Phase: api.PodSucceeded}

	podList := api.PodList{}
	podList.Items = []api.Pod{
		podPending,
		podFailed,
		podUnknown,
		podRunning,
		podSucceeded,
	}

	found := KubePodsFilter{podList}.ByService("web").ByStatus("Running").ByStatusReason("Running").First()
	assert.Equal(t, "web-812374193-mxiwy", found.Name)
}
