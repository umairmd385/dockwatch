package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Client struct {
	api client.APIClient
}

// NewClient creates a new Docker client connecting to the local daemon
func NewClient() (*Client, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return &Client{api: apiClient}, nil
}

// ListContainers returns all running containers
func (c *Client) ListContainers(ctx context.Context) ([]types.Container, error) {
	return c.api.ContainerList(ctx, container.ListOptions{All: false})
}
