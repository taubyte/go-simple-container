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
func (c *Client) Image(ctx context.Context, image string, options ...ImageOption) (*Image, error) {
	_image := &Image{
		client: c,
		image:  image,
	}

	for _, opt := range options {
		err := opt(_image)
		if err != nil {
			return nil, fmt.Errorf("running image options for image `%s` failed with: %s", image, err)
		}
	}

	imageExists := _image.checkImageExists(ctx)
	if _image.buildTarball != nil {
		if ForceRebuild || !imageExists {
			err := _image.buildImage(ctx)
			if err != nil {
				return nil, fmt.Errorf("building image `%s` failed with %s", image, err)
			}
		}

		return _image, nil
	}

	_image, err := _image.Pull(ctx)
	if err != nil && !imageExists {
		return nil, fmt.Errorf("pulling image `%s` failed with %s", image, err)
	}

	return _image, nil
}

// checkImage checks the docker host client if the image is known.
func (i *Image) checkImageExists(ctx context.Context) bool {
	res, err := i.client.ImageList(ctx, types.ImageListOptions{
		Filters: NewFilter("reference", i.image),
	})
	if err != nil || len(res) < 1 {
		return false
	}

	return true
}

// buildImage builds a DockerFile tarball as a docker image.
func (i *Image) buildImage(ctx context.Context) error {
	imageBuildResponse, err := i.client.ImageBuild(
		ctx,
		i.buildTarball,
		types.ImageBuildOptions{
			Context:    i.buildTarball,
			Dockerfile: "Dockerfile",
			Tags:       []string{i.image},
			Remove:     true})
	if err != nil {
		return fmt.Errorf("building Dockerfile for image `%s` failed with: %s", i.image, err)
	}

	defer imageBuildResponse.Body.Close()

	_, err = io.Copy(os.Stdout, imageBuildResponse.Body)
	if err != nil {
		return fmt.Errorf("copying response from  image `%s` build failed with :%s", i.image, err)
	}

	return nil
}

// Pull retrieves latest changes to the image from docker hub.
func (i *Image) Pull(ctx context.Context) (*Image, error) {
	reader, err := i.client.ImagePull(ctx, i.image, types.ImagePullOptions{})
	if err != nil {
		return i, fmt.Errorf("pulling image `%s`, failed with %s", i.image, err)
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
		err = json.Unmarshal(line, &status)
		if err != nil {
			return i, fmt.Errorf("unmarshaling status from docker pull on image `%s` failed with: %s", i.image, err)
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
			return nil, fmt.Errorf("running image options for image `%s` failed with: %s", i.image, err)
		}
	}

	var mounts []mount.Mount
	_volume := mount.Mount{
		Type: mount.TypeBind,
	}
	for _, volume := range c.volumes {
		_volume.Source = volume.source
		_volume.Target = volume.target
		mounts = append(mounts, _volume)
	}

	config := &container.Config{
		Image: i.image,
		Cmd:   c.cmd,
		Shell: c.shell,
		Tty:   false,
		Env:   c.env,
	}

	if len(c.workDir) != 0 {
		config.WorkingDir = c.workDir
	}

	resp, err := c.image.client.ContainerCreate(ctx, config, &container.HostConfig{Mounts: mounts}, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("creating container for image `%s` failed with: %s", c.image.image, err)
	}
	c.id = resp.ID

	return c, nil
}

// Clean removes all docker images that match the given filter, and max age.
func (c *Client) Clean(ctx context.Context, age time.Duration, filter filters.Args) error {
	images, _ := c.ImageList(context.Background(), types.ImageListOptions{Filters: filter})
	timeNow := time.Now()
	for _, image := range images {
		if time.Unix(image.Created, 0).Add(age).Before(timeNow) {
			_, err := c.ImageRemove(ctx, image.ID, types.ImageRemoveOptions{
				Force:         true,
				PruneChildren: true,
			})
			if err != nil {
				return fmt.Errorf("cleaning image `%s` from docker host failed with %s", image.ID, err)
			}
		}
	}

	return nil
}

// Name returns the name of the image
func (i *Image) Name() string {
	return i.image
}
