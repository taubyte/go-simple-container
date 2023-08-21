package containers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
)

// Image initializes the given image, and attempts to pull the container from docker hub.
// If the Build() Option is provided then the given DockerFile tarball is built and returned.
func (c *Client) Image(ctx context.Context, name string, options ...ImageOption) (image *Image, err error) {
	image = &Image{
		client: c,
		image:  name,
	}

	for _, opt := range options {
		if err := opt(image); err != nil {
			return nil, errorImageOptions(name, err)
		}
	}

	imageExists := image.checkImageExists(ctx)
	if image.buildTarball != nil && (ForceRebuild || !imageExists) {
		if err := image.buildImage(ctx); err != nil {
			return nil, errorImageBuild(name, err)
		}
	} else {
		if image, err = image.Pull(ctx); err != nil {
			return nil, errorImagePull(name, err)
		}
	}

	return
}

// checkImage checks the docker host client if the image is known.
func (i *Image) checkImageExists(ctx context.Context) bool {
	res, err := i.client.Docker.ImageList(ctx, types.ImageListOptions{
		Filters: NewFilter("reference", i.image),
	})

	return err == nil && len(res) > 0
}

// buildImage builds a DockerFile tarball as a docker image.
func (i *Image) buildImage(ctx context.Context) error {
	imageBuildResponse, err := i.client.Docker.ImageBuild(
		ctx,
		i.buildTarball,
		types.ImageBuildOptions{
			Context:    i.buildTarball,
			Dockerfile: "Dockerfile",
			Tags:       []string{i.image},
			Remove:     true,
		},
	)
	if err != nil {
		return errorImageBuildDockerFile(err)
	}
	defer imageBuildResponse.Body.Close()

	if _, err = io.Copy(os.Stdout, imageBuildResponse.Body); err != nil {
		return errorImageBuildResCopy(err)
	}

	return nil
}

// Pull retrieves latest changes to the image from docker hub.
func (i *Image) Pull(ctx context.Context) (*Image, error) {
	reader, err := i.client.Docker.ImagePull(ctx, i.image, types.ImagePullOptions{})
	if err != nil {
		return i, errorClientPull(err)
	}
	defer reader.Close()

	fileScanner := bufio.NewScanner(reader)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		var line = fileScanner.Bytes()
		var status struct {
			Status         string
			ProgressDetail struct {
				Current int
				Total   int
			}
			Id string
		}

		if err = json.Unmarshal(line, &status); err != nil {
			return i, errorImagePullStatus(err)
		}
	}

	return i, nil
}

// Instantiate sets given options and creates the container from the docker image.
func (i *Image) Instantiate(ctx context.Context, options ...ContainerOption) (*Container, error) {
	c := &Container{
		image: i,
	}
	for _, opt := range options {
		err := opt(c)
		if err != nil {
			return nil, errorContainerOptions(i.image, err)
		}
	}

	mounts := make([]mount.Mount, len(c.volumes))
	for idx, volume := range c.volumes {
		mounts[idx] = mount.Mount{
			Type:   mount.TypeBind,
			Source: volume.source,
			Target: volume.target,
		}
	}

	config := &container.Config{
		Image: i.image,
		Cmd:   c.cmd,
		Shell: c.shell,
		Tty:   false,
		Env:   c.env,
	}
	if len(c.workDir) > 0 {
		config.WorkingDir = c.workDir
	}

	resp, err := c.image.client.Docker.ContainerCreate(ctx, config, &container.HostConfig{Mounts: mounts}, nil, nil, "")
	if err != nil {
		return nil, errorContainerCreate(c.image.Name(), err)
	}
	c.id = resp.ID

	return c, nil
}

// Clean removes all docker images that match the given filter, and max age.
func (c *Client) Clean(ctx context.Context, age time.Duration, filter filters.Args) error {
	images, _ := c.Docker.ImageList(ctx, types.ImageListOptions{Filters: filter})
	timeNow := time.Now()

	var err error
	for _, image := range images {
		if time.Unix(image.Created, 0).Add(age).Before(timeNow) {
			if _, _err := c.Docker.ImageRemove(ctx, image.ID, types.ImageRemoveOptions{
				Force:         true,
				PruneChildren: true,
			}); _err != nil {
				if err != nil {
					err = fmt.Errorf("%s:%w", err, _err)
				} else {
					err = _err
				}
			}
		}
	}

	return err
}

// Name returns the name of the image
func (i *Image) Name() string {
	return i.image
}
