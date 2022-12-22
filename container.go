package containers

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

// Run starts the container and waits for the container to exit before returning the container logs.
func (c *Container) Run(ctx context.Context) (*MuxedReadCloser, error) {
	if err := c.image.client.ContainerStart(ctx, c.id, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("starting container with id `%s` for image `%s` failed with: %s", c.id, c.image.image, err)
	}

	err := c.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("waiting for container with id `%s` for image `%s` failed with: %s", c.id, c.image.image, err)
	}

	info, err := c.image.client.ContainerInspect(ctx, c.id)
	if err != nil {
		return nil, fmt.Errorf("inspecting container with id `%s` for image `%s` failed with: %s", c.id, c.image.image, err)
	}

	var RetCodeErr error
	if info.ContainerJSONBase.State.ExitCode != 0 {
		RetCodeErr = fmt.Errorf("running container with id `%s` for image `%s` failed with exit-code: %d", c.id, c.image.image, info.ContainerJSONBase.State.ExitCode)
	}

	muxed, err := c.image.client.ContainerLogs(ctx, c.id, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return nil, fmt.Errorf("getting container logs for image `%s` failed with %s", c.image.image, err)
	}

	c.Cleanup(ctx)

	return &MuxedReadCloser{reader: muxed}, RetCodeErr
}

// Wait calls the ContainerWait method for the container, and returns once a response has been received.
// If there is an error response then wait will return the error
func (c *Container) Wait(ctx context.Context) error {
	statusCh, errCh := c.image.client.ContainerWait(ctx, c.id, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("container with id `%s` returned error: %s", c.id, err)
		}
	case <-statusCh:
	}
	return nil
}

// Cleanup removes the container from the docker host client.
func (c *Container) Cleanup(ctx context.Context) error {
	if err := c.image.client.ContainerRemove(ctx, c.id, types.ContainerRemoveOptions{}); err != nil {
		return fmt.Errorf("removing container with id `%s` failed with: %s", c.id, err)
	}
	return nil
}
