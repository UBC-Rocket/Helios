package docker

import (
	"fmt"

	"github.com/docker/docker/api/types/network"
)

// Start up a Docket network with the provided name.
// If it already exists, it finds and returns the ID of the existing network.
// Network will be saved into the client object.
// Returns the network and any errors
func (c *DockerClient) InitializeNetwork(networkName string) error {
	list, netErr := c.cli.NetworkList(c.ctx, network.ListOptions{})
	if netErr != nil {
		return netErr
	}

	for _, net := range list {
		if net.Name == networkName {
			fmt.Println("'", networkName, "' network already exists:", net.ID)
			c.net = network.CreateResponse{ID: net.ID}
			return nil
		}
	}

	// Create network
	fmt.Println("Creating Docker network '", networkName, "'...")
	result, err := c.cli.NetworkCreate(c.ctx, networkName, network.CreateOptions{
		Driver: "bridge",
	})
	if err != nil {
		return err
	}
	fmt.Println("'", networkName, "' network created:", result.ID)

	c.net = result
	return nil
}

// Add a container to the specified Docker Network
func (c *DockerClient) AddContainerToNetwork(containerID string) error {
	err := c.cli.NetworkConnect(c.ctx, c.net.ID, containerID, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Container %s connected to docker network\n", containerID)
	return nil
}
