package docker

import (
	"fmt"

	"github.com/filipecaixeta/logviewer/internal/source"

	"github.com/docker/docker/api/types"
)

// ContainerItem represents a Docker container in the list.
type ContainerItem struct {
	ID     string
	Name   string
	Image  string
	Status string
}

// NewContainerItem creates a new instance of ContainerItem from a Docker API Container type.
func NewContainerItem(container types.Container) *ContainerItem {
	return &ContainerItem{
		ID:     container.ID,
		Name:   getContainerName(container.Names),
		Image:  container.Image,
		Status: container.Status,
	}
}

// String returns a string representation of the container, which will be displayed in the list.
func (c *ContainerItem) String() string {
	return fmt.Sprintf("%s | %s", c.Name, c.Image)
}

// Children returns an empty slice, as Docker containers don't have a hierarchical structure in this context.
func (c *ContainerItem) Children() []source.ListItem {
	return []source.ListItem{}
}

// FilterValue returns the value used for filtering the list. It could be the container's name or ID.
func (c *ContainerItem) FilterValue() string {
	return c.Name
}

// getContainerName safely retrieves the container's name from the list of names.
func getContainerName(names []string) string {
	if len(names) > 0 {
		return names[0][1:] // Docker container names come with a '/' prefix, you may choose to remove it.
	}
	return "Unknown"
}
