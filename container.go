package main

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"io"
	"net/http"
	"os"
	"strings"
)

var c *ClientAPI

type ClientAPI struct {
	DockerCli *client.Client
	Ctx       context.Context
}

type ClientAPIer interface {
	Stop(containerId string, timeout int) error
	Restart(containerId string, timeout int) error
	PullImage(imageUrl string) error
	Create(config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (string, error)
	Start(containerId string, startOpts types.ContainerStartOptions) error
	EnsureImageExist(image string) (bool, error)
	CreateNetwork(containerName string) (string, error)
	ConnectNetwork(networkId string, containerId string) error
	RemoveImage(imageId string, removeOpt types.ImageRemoveOptions) error
	Status(containerId string) (*types.Container, error)
	Inspect(containerId string) (*types.ContainerJSON, error)
	Stats(containerId string, stream bool) (*types.ContainerStats, error)
	RemoveContainer(containerId string, removeVolumes bool) error
	RemoveNetwork(networkId string) error
	Exec(containerId string, cmd []string) error
	EnsureNetworkExist(networkName string) (string, error)
	FindContainer(containerName string) (string, error)
	IsImageUsed(imageID string) (bool, error)
	ContainerLogs(writer http.ResponseWriter, containerId string, follow bool, since string, tail string, until string) error
	ExportImage(imageId string) (io.ReadCloser, error)
	ImportImage(savePath string, repo string, tag string) error
	ImageInfo(imageId string) (*types.ImageInspect, error)
}

func GetContainerAPIer() ClientAPIer {
	return c
}

func NewContainerMgr(apiVersion string) ClientAPIer {
	return &ClientAPI{
		DockerCli: DockerClient(apiVersion),
		Ctx:       context.Background(),
	}
}

func (c *ClientAPI) Create(config *container.Config, hostConfig *container.HostConfig, networkConfig *network.NetworkingConfig, containerName string) (string, error) {
	//var networkConfig *network.NetworkingConfig

	rspBody, err := c.DockerCli.ContainerCreate(c.Ctx, config, hostConfig, networkConfig, nil, containerName)
	if err != nil {
		return "", err
	}
	return rspBody.ID, nil
}

func (c *ClientAPI) Start(containerId string, startOpts types.ContainerStartOptions) error {
	err := c.DockerCli.ContainerStart(c.Ctx, containerId, startOpts)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClientAPI) Status(containerId string) (*types.Container, error) {
	var args = filters.NewArgs()
	args.Add("id", containerId)
	containerInfo, err := c.DockerCli.ContainerList(c.Ctx, types.ContainerListOptions{
		All:     true,
		Filters: args})

	if err != nil {
		return nil, err
	}
	if len(containerInfo) > 0 {
		return &containerInfo[0], nil
	}
	return nil, errors.New("not found container")
}

func (c *ClientAPI) Inspect(containerId string) (*types.ContainerJSON, error) {
	containerInpsect, err := c.DockerCli.ContainerInspect(c.Ctx, containerId)
	if err != nil {
		return nil, err
	}
	return &containerInpsect, nil
}

func (c *ClientAPI) Stats(containerId string, stream bool) (*types.ContainerStats, error) {
	containerStats, err := c.DockerCli.ContainerStats(c.Ctx, containerId, stream)
	if err != nil {
		return nil, err
	}
	return &containerStats, nil
}

func (c *ClientAPI) Stop(containerId string, timeout int) error {
	if timeout == 0 {
		timeout = 10
	}
	options := container.StopOptions{
		Signal:  "",
		Timeout: &timeout,
	}

	err := c.DockerCli.ContainerStop(c.Ctx, containerId, options)

	if err != nil {
		return err
	}
	return nil
}

func (c *ClientAPI) Restart(containerId string, timeout int) error {
	if timeout == 0 {
		timeout = 10
	}
	options := container.StopOptions{
		Signal:  "",
		Timeout: &timeout,
	}

	err := c.DockerCli.ContainerRestart(c.Ctx, containerId, options)

	if err != nil {
		return err
	}
	return nil
}

func (c *ClientAPI) RemoveContainer(containerId string, removeVolumes bool) error {
	//var duration = 10 * time.Second

	err := c.DockerCli.ContainerRemove(c.Ctx, containerId, types.ContainerRemoveOptions{RemoveVolumes: removeVolumes})

	if err != nil {
		return err
	}
	return nil
}

func (c *ClientAPI) Exec(containerId string, cmd []string) error {
	execId, err := c.DockerCli.ContainerExecCreate(c.Ctx, containerId, types.ExecConfig{
		Cmd: cmd,
	})
	if err != nil {
		return err
	}
	//err = c.DockerCli.ContainerExecStart(c.Ctx, execId.ID, types.ExecStartCheck{
	//	Detach: false,
	//	Tty:    false,
	//})
	//if err != nil {
	//	return err
	//}
	resp, err := c.DockerCli.ContainerExecAttach(c.Ctx, execId.ID, types.ExecStartCheck{
		Detach: false,
		Tty:    false,
	})
	defer resp.Close()
	io.Copy(os.Stdout, resp.Reader)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClientAPI) FindContainer(containerName string) (string, error) {
	var args = filters.NewArgs()
	args.Add("name", containerName)
	containerInfos, err := c.DockerCli.ContainerList(c.Ctx, types.ContainerListOptions{
		All:     true,
		Filters: args})

	if err != nil {
		return "", err
	}
	if len(containerInfos) > 0 {
		for i, ci := range containerInfos {
			if strings.TrimLeft(ci.Names[0], "/") == containerName {
				return containerInfos[i].ID, nil
			}
		}
	}
	return "", nil
}

func (c *ClientAPI) ContainerLogs(writer http.ResponseWriter, containerId string, follow bool, since string, tail string, until string) error {
	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      since,
		Until:      until,
		Timestamps: false,
		Follow:     follow,
		Tail:       tail,
		Details:    false,
	}
	responseBody, err := c.DockerCli.ContainerLogs(c.Ctx, containerId, options)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	writer.Header().Add("Content-Type", "application/octet-stream")
	writer.WriteHeader(200)
	_, err = io.Copy(writer, responseBody)
	return err
}
