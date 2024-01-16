package k8s

import (
	"context"
	"sync"

	"github.com/filipecaixeta/logviewer/internal/source"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Namespace struct {
	Name      string      `json:"name"`
	Workloads []*Workload `json:"workloads"`
}

func (n *Namespace) Children() []source.ListItem {
	return source.Convert2ListItems(n.Workloads)
}

func (n *Namespace) String() string {
	return n.Name
}

func (n *Namespace) FilterValue() string {
	return n.Name
}

func NewNamespace(name string, clientset *kubernetes.Clientset) *Namespace {
	var replicaSets *appsv1.ReplicaSetList
	var namespacedPods *v1.PodList
	var deployments *appsv1.DeploymentList
	var statefulsets *appsv1.StatefulSetList
	var wg sync.WaitGroup

	var n = &Namespace{Name: name}

	n.Workloads = make([]*Workload, 0)

	wg.Add(4)
	go func() {
		replicaSets, _ = clientset.AppsV1().ReplicaSets(n.Name).List(context.Background(), metav1.ListOptions{})
		wg.Done()
	}()
	go func() {
		namespacedPods, _ = clientset.CoreV1().Pods(n.Name).List(context.Background(), metav1.ListOptions{})
		wg.Done()
	}()
	go func() {
		deployments, _ = clientset.AppsV1().Deployments(n.Name).List(context.Background(), metav1.ListOptions{})
		wg.Done()
	}()
	go func() {
		statefulsets, _ = clientset.AppsV1().StatefulSets(n.Name).List(context.Background(), metav1.ListOptions{})
		wg.Done()
	}()
	wg.Wait()

	for _, deployment := range deployments.Items {
		workload := &Workload{
			Kind:              "Deployment",
			Name:              deployment.Name,
			UID:               string(deployment.UID),
			Labels:            deployment.Labels,
			CreationTimestamp: deployment.CreationTimestamp.Time,
			Namespace:         deployment.Namespace,
		}
		workload.Pods = GetPodsForWorkload(replicaSets.Items, &deployment, namespacedPods.Items, "Deployment")
		n.Workloads = append(n.Workloads, workload)
	}

	for _, statefulset := range statefulsets.Items {
		workload := &Workload{
			Kind:              "StatefulSet",
			Name:              statefulset.Name,
			UID:               string(statefulset.UID),
			Labels:            statefulset.Labels,
			CreationTimestamp: statefulset.CreationTimestamp.Time,
			Namespace:         statefulset.Namespace,
		}
		workload.Pods = GetPodsForWorkload(replicaSets.Items, &statefulset, namespacedPods.Items, "StatefulSet")
		n.Workloads = append(n.Workloads, workload)
	}

	return n
}
