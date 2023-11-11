package main

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
)

func (c *ClientAPI) CreateNetwork(containerName string) (string, error) {
	createdNetwork, err := c.DockerCli.NetworkCreate(c.Ctx, containerName, types.NetworkCreate{})
	if err != nil {
		return "", err
	}
	return createdNetwork.ID, nil
}

func (c *ClientAPI) ConnectNetwork(networkId string, containerId string) error {
	var endpointSettings *network.EndpointSettings

	err := c.DockerCli.NetworkConnect(c.Ctx, networkId, containerId, endpointSettings)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClientAPI) RemoveNetwork(networkId string) error {
	err := c.DockerCli.NetworkRemove(c.Ctx, networkId)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClientAPI) EnsureNetworkExist(networkName string) (string, error) {
	var args = filters.NewArgs()
	args.Add("name", networkName)
	networkRes, err := c.DockerCli.NetworkList(c.Ctx, types.NetworkListOptions{Filters: args})
	if err != nil {
		return "", err
	}
	if len(networkRes) > 0 {
		for _, network := range networkRes {
			if network.Name == networkName {
				return network.ID, nil
			}
		}
		return "", nil
	} else {
		return "", nil
	}
}
