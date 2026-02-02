package service

import (
	"context"
	"fmt"
	"io"

	"ServerMonitor/internal/model"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var dockerCli *client.Client

// InitDockerClient initializes the global docker client
func InitDockerClient() error {
	var err error
	dockerCli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	return err
}

// ListContainers returns a list of all containers
func ListContainers() ([]model.ContainerSimple, error) {
	containers, err := dockerCli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	var list []model.ContainerSimple
	for _, c := range containers {
		name := "unknown"
		if len(c.Names) > 0 {
			name = c.Names[0]
		}
		// Short ID
		shortID := c.ID
		if len(shortID) > 10 {
			shortID = shortID[:10]
		}
		list = append(list, model.ContainerSimple{
			ID:     shortID,
			Name:   name,
			Image:  c.Image,
			Status: c.Status,
			State:  c.State,
		})
	}
	return list, nil
}

// PerformAction executes start, stop, or restart on a container
func PerformAction(id, action string) error {
	ctx := context.Background()
	var err error

	switch action {
	case "start":
		err = dockerCli.ContainerStart(ctx, id, container.StartOptions{})
	case "stop":
		err = dockerCli.ContainerStop(ctx, id, container.StopOptions{})
	case "restart":
		err = dockerCli.ContainerRestart(ctx, id, container.StopOptions{})
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	return err
}

// GetLogs returns a ReadCloser stream of logs
func GetLogs(id string) (io.ReadCloser, error) {
	ctx := context.Background()
	options := container.LogsOptions{ShowStdout: true, ShowStderr: true, Tail: "100"}
	return dockerCli.ContainerLogs(ctx, id, options)
}
