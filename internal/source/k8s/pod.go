package k8s

import (
	"strings"
	"time"

	"github.com/filipecaixeta/logviewer/internal/source"

	v1 "k8s.io/api/core/v1"
)

type Pod struct {
	Name              string            `json:"name"`
	UID               string            `json:"uid"`
	Labels            map[string]string `json:"-"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Namespace         string            `json:"namespace"`
	Status            string            `json:"status"`
	Containers        []*Container      `json:"containers"`
}

func (n *Pod) Children() []source.ListItem {
	return source.Convert2ListItems(n.Containers)
}

func (n *Pod) String() string {
	return n.Name
}

func (n *Pod) FilterValue() string {
	return n.Name
}

func NewPod(pod v1.Pod) *Pod {
	containers := make([]*Container, len(pod.Spec.Containers))
	for i, c := range pod.Spec.Containers {
		containers[i] = &Container{
			Name:     c.Name,
			Image:    c.Image,
			ImageTag: c.Image[strings.Index(c.Image, ":")+1:],
		}
	}

	return &Pod{
		Name:              pod.Name,
		UID:               string(pod.UID),
		Labels:            pod.Labels,
		CreationTimestamp: pod.CreationTimestamp.Time,
		Namespace:         pod.Namespace,
		Status:            string(pod.Status.Phase),
		Containers:        containers,
	}
}
