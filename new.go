package containers

import (
	"fmt"

	"github.com/docker/docker/client"
)

// New creates a new dockerClient with default Options.
func New() (dockerClient *Client, err error) {
	dockerClient = new(Client)
	dockerClient.Client, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("new docker client failed with: %w", err)
	}

	return
}
