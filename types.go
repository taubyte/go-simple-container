package containers

import (
	"io"

	"github.com/docker/docker/client"
)

var (
	ForceRebuild = false
)

// MuxedReadCloser wraps the Read/Close methods for muxed logs.
type MuxedReadCloser struct {
	reader io.ReadCloser
}

// Client wraps the methods of the docker Client.
type Client struct {
	*client.Client
}

// volume defines the source and target to be volumed in the docker container.
type volume struct {
	source string
	target string
}

// Container wraps the methods of the docker container.
type Container struct {
	image   *Image
	id      string
	cmd     []string
	shell   []string
	volumes []volume
	env     []string
	workDir string
}

// Image wraps the methods of the docker image.
type Image struct {
	client       *Client
	image        string
	buildTarball io.Reader
}
