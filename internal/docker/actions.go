package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
)

// Restart restarts a container gracefully
func (c *Client) Restart(ctx context.Context, containerID string) error {
	timeout := 10 // seconds
	return c.api.ContainerRestart(ctx, containerID, container.StopOptions{Timeout: &timeout})
}

// Signal sends a UNIX signal to a container
func (c *Client) Signal(ctx context.Context, containerID string, signal string) error {
	return c.api.ContainerKill(ctx, containerID, signal)
}

// ExecHTTP executes an HTTP request inside the container to trigger a reload endpoint
func (c *Client) ExecHTTP(ctx context.Context, containerID string, endpoint string) error {
	// Attempt curl first, fallback to wget if not available
	cmdStr := fmt.Sprintf("curl -X POST -s %s || wget -qO- --post-data='' %s", endpoint, endpoint)
	cmd := []string{"sh", "-c", cmdStr}

	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execIDResp, err := c.api.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	err = c.api.ContainerExecStart(ctx, execIDResp.ID, container.ExecStartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start exec: %w", err)
	}

	return nil
}
