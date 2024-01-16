package docker

import (
	"bufio"
	"context"

	"github.com/filipecaixeta/logviewer/internal/source"
	"github.com/filipecaixeta/logviewer/internal/state"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type DockerSource struct {
	columns   []*source.List
	dockerCli *client.Client
}

func New() source.Source {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return &DockerSource{
		columns: []*source.List{
			source.NewList("Containers", []source.ListItem{}),
		},
		dockerCli: cli,
	}
}

func (d *DockerSource) Init(stateChan chan state.State) tea.Cmd {
	return func() tea.Msg {
		d.refreshContainers()
		stateChan <- state.StateBrose
		return nil
	}
}

func (d *DockerSource) Columns() []*source.List {
	return d.columns
}

func (d *DockerSource) Logs(ctx context.Context, stateChan chan state.State, logChan chan string) tea.Cmd {
	return func() tea.Msg {
		selectedItem := d.columns[0].SelectedItem()
		if selectedItem == nil {
			return nil // Or handle the no selection case
		}

		container, ok := selectedItem.(*ContainerItem)
		if !ok {
			return nil // Or handle the type assertion error
		}

		containerID := container.ID

		options := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true}
		logsReader, err := d.dockerCli.ContainerLogs(ctx, containerID, options)
		if err != nil {
			// Handle error
			return err
		}
		defer logsReader.Close()

		scanner := bufio.NewScanner(logsReader)
		stateChan <- state.StateLogs
		for scanner.Scan() {
			t := scanner.Text()
			// remove log line header
			if len(t) > 8 {
				logChan <- t[8:]
			}
		}

		if err := scanner.Err(); err != nil {
			// Handle scanning error
			return err
		}

		return nil
	}
}

func (d *DockerSource) refreshContainers() {
	containers, err := d.dockerCli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		// Handle error
		return
	}

	containerItems := []source.ListItem{}
	for _, container := range containers {
		containerItems = append(containerItems, NewContainerItem(container))
	}

	d.columns[0].SetItems(containerItems)
}
