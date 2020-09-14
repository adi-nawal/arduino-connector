package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestInstallUbuntuDockerPath(t *testing.T) {
	value, err := isDockerInstalled()
	assert.False(t, value)
	assert.True(t, err == nil)
}

func TestInstallAlpineDockerPath(t *testing.T) {
	value, err := isDockerInstalled()
	assert.True(t, value)
	assert.True(t, err == nil)
}

func TestInstallAlpineDockerImagesEmpty(t *testing.T) {
	values, err := retrieveDockerImages()
	assert.True(t, err == nil)
	assert.True(t, len(values) == 0)
	assert.True(t, reflect.DeepEqual(values, []string{}))
}

func TestInstallAlpineDockerImages(t *testing.T) {
	cli, err := docker.NewClientWithOpts(docker.WithVersion("1.38"))
	assert.True(t, err == nil)

	ctx := context.Background()
	_, err = cli.ImagePull(ctx, "alpine", types.ImagePullOptions{})
	time.Sleep(5 * time.Second)
	assert.True(t, err == nil)

	defer func() {
		forceAllImagesArg, _ := filters.FromJSON(`{"dangling": false}`)
		_, err = cli.ImagesPrune(ctx, forceAllImagesArg)
		assert.True(t, err == nil)
	}()

	values, err := retrieveDockerImages()
	assert.True(t, err == nil)
	assert.True(t, len(values) == 1)
	assert.True(t, reflect.DeepEqual(values, []string{"alpine:latest"}))
}

func TestInstallAlpineDockerContainersEmpty(t *testing.T) {
	values, err := retrieveDockerContainer()
	assert.True(t, err == nil)
	assert.True(t, len(values) == 0)
	assert.True(t, reflect.DeepEqual(values, []string{}))
}

func TestInstallAlpineDockerContainers(t *testing.T) {
	cli, err := docker.NewClientWithOpts(docker.WithVersion("1.38"))
	assert.True(t, err == nil)

	ctx := context.Background()
	_, err = cli.ImagePull(ctx, "alpine", types.ImagePullOptions{})
	time.Sleep(5 * time.Second)
	assert.True(t, err == nil)

	defer func() {
		filters := filters.NewArgs(filters.Arg("reference", "alpine"))
		images, errImagels := cli.ImageList(context.Background(), types.ImageListOptions{Filters: filters})
		assert.True(t, errImagels == nil)
		_, err = cli.ImageRemove(ctx, images[0].ID, types.ImageRemoveOptions{})
		time.Sleep(5 * time.Second)
		assert.True(t, err == nil)
	}()

	createContResp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   []string{"echo", "hello world"},
	}, nil, nil, "")
	time.Sleep(5 * time.Second)
	assert.True(t, err == nil)

	defer func() {
		err = cli.ContainerRemove(ctx, createContResp.ID, types.ContainerRemoveOptions{})
		time.Sleep(5 * time.Second)
		assert.True(t, err == nil)
	}()

	defer func() {
		forceAllImagesArg, _ := filters.FromJSON(`{"dangling": false}`)
		_, err = cli.ImagesPrune(ctx, forceAllImagesArg)
		time.Sleep(5 * time.Second)
		assert.True(t, err == nil)
	}()

	values, err := retrieveDockerContainer()
	assert.True(t, err == nil)
	assert.True(t, len(values) == 1)
	assert.True(t, reflect.DeepEqual(values, []string{createContResp.ID}))
}

func TestInstallUbuntuNetworkManager(t *testing.T) {
	assert.False(t, isNetManagerInstalled())
}
