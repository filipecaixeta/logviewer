package k8s

import (
	"time"

	"github.com/filipecaixeta/logviewer/internal/source"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Workload struct {
	Kind              string            `json:"kind"`
	Name              string            `json:"name"`
	UID               string            `json:"uid"`
	Labels            map[string]string `json:"-"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Namespace         string            `json:"namespace"`
	Pods              []*Pod            `json:"pods"`
}

func (n *Workload) Children() []source.ListItem {
	return source.Convert2ListItems(n.Pods)
}

func (n *Workload) String() string {
	return n.Name
}

func (n *Workload) FilterValue() string {
	return n.Name
}

// GetPodsForWorkload finds and returns a list of pods associated with a given workload.
// The workload can be either a Deployment or a StatefulSet.
//
// Parameters:
// - replicaSets: A slice of ReplicaSet objects in the namespace.
// - workload: The specific workload object (Deployment or StatefulSet).
// - namespacedPods: A slice of Pod objects in the namespace.
// - workloadKind: A string indicating the type of the workload ("Deployment" or "StatefulSet").
//
// Returns:
// - A slice of Pod objects that are associated with the given workload.
func GetPodsForWorkload(replicaSets []appsv1.ReplicaSet, workload metav1.Object, namespacedPods []v1.Pod, workloadKind string) []*Pod {
	var podsForWorkload []*Pod
	workloadUID := string(workload.GetUID())

	switch workloadKind {
	case "Deployment":
		// Find all ReplicaSets owned by the Deployment.
		var ownedReplicaSets []appsv1.ReplicaSet
		for _, rs := range replicaSets {
			for _, owner := range rs.OwnerReferences {
				if string(owner.UID) == workloadUID {
					ownedReplicaSets = append(ownedReplicaSets, rs)
				}
			}
		}

		// Find all Pods owned by these ReplicaSets.
		for _, rs := range ownedReplicaSets {
			rsUID := string(rs.UID)
			for _, pod := range namespacedPods {
				for _, owner := range pod.OwnerReferences {
					if string(owner.UID) == rsUID {
						podsForWorkload = append(podsForWorkload, NewPod(pod))
					}
				}
			}
		}
	case "StatefulSet":
		// Find all Pods owned by the StatefulSet.
		for _, pod := range namespacedPods {
			for _, owner := range pod.OwnerReferences {
				if string(owner.UID) == workloadUID {
					podsForWorkload = append(podsForWorkload, NewPod(pod))
				}
			}
		}
	default:
		// if the workload kind is not supported.
		return nil
	}

	return podsForWorkload
}
