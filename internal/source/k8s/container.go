package k8s

import "github.com/filipecaixeta/logviewer/internal/source"

type Container struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	ImageTag string `json:"imageTag"`
}

func (n *Container) Children() []source.ListItem {
	return nil
}

func (n *Container) String() string {
	return n.Name
}

func (n *Container) FilterValue() string {
	return n.Name
}
