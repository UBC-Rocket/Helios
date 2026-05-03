package docker

import (
	"context"
	"fmt"
	"strings"

	componenttree "helios/internal/component_tree"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type DockerClient struct {
	runtimeHash string
	cli *client.Client
	ctx context.Context
	net network.CreateResponse
}

/*
 * ========================================================
 * = Public methods for managing Docker runtime
 * ========================================================
 */

func NewDockerClient(ctx context.Context, hash string) *DockerClient {
	return &DockerClient{
		ctx:         ctx,
		runtimeHash: hash,
	}
}

// Initialize the Docker client.
func (c *DockerClient) Initialize() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	c.cli = cli
	return nil
}

// Start configured components in Docker based on the component tree in the core.
func (c *DockerClient) StartConfigured(t *componenttree.ComponentTree) error {
	if err := t.EachComponent(c.startContainer); err != nil {
		return err
	}
	return nil
}

// Close the Docker client.
func (c *DockerClient) Close() error {
	return c.cli.Close()
}


/*
 * ========================================================
 * = Internal structures for managing component state and connections
 * ========================================================
 */


// Get the ID of an existing container given it's name.
// Returns the ID if found and "" if it does not exist.
func (c *DockerClient) getContainerID(containerName string) (containerID string) {
	list := c.getContainers()
	var contID string = ""

	for _, cont := range list {
		if cont.Names[0] == "/"+containerName {
			contID = cont.ID
			break
		}
	}

	return contID
}

// Get a list of all containers.
func (c *DockerClient) getContainers() (summary []container.Summary) {
	list, err := c.cli.ContainerList(c.ctx, container.ListOptions{All: true})
	if err != nil {
		panic(err)
	}
	return list
}

// Start a container given a pointer to the component struct.
// If the container already exists, it will be restarted or removed and recreated if the hash does not match.
// Created container will be added to the docker network and started.
func (c *DockerClient) startContainer(address string, name string, comp *componenttree.Component) error {
	list := c.getContainers()
	var cont container.Summary = container.Summary{}

	// Find container by name to see if it exists
	for _, x := range list {
		if x.Names[0] == "/"+name {
			cont = x
			break
		}
	}

	// If container exists, handle the various cases
	if (cont.ID != "") {
		// If the runtime hash does not match
		if cont.Labels["runtime_hash"] != c.runtimeHash {
			// Remove the container
			if err := c.cli.ContainerRemove(c.ctx, cont.ID, container.RemoveOptions{Force: true}); err != nil {
				return err
			}

			// Reset cont to indicate it does not exist
			cont = container.Summary{}
		
		// If the runtime hash matches, check if is exited or running
		} else {
			if cont.State == "running" {
				// Do nothing
				comp.SetDockerConnID(cont.ID)
				return nil
			} else if cont.State == "exited" {
				// Restart it
				if err := c.startDockerContainer(cont.ID); err != nil {
					return err
				}
				comp.SetDockerConnID(cont.ID)
				return nil
			} else if cont.State == "created" {
				// Start it
				if err := c.startDockerContainer(cont.ID); err != nil {
					return err
				}
				comp.SetDockerConnID(cont.ID)
				return nil
			}
		}
	}

	// Container does not exist, create it
	contResp, contErr := c.createContainer(address, name, comp)
	if contErr != nil {
		return contErr
	}
	cont.ID = contResp.ID

	comp.SetDockerConnID(cont.ID)

	// Add to network and start container
	if err := c.AddContainerToNetwork(cont.ID); err != nil {
		return err
	}
	if err := c.startDockerContainer(cont.ID); err != nil {
		return err
	}
	return nil
}

// Create a container using information from the image struct and runtime_hash.
// It should be checked if a container already exists with the same name and hash before calling this function.
func (c *DockerClient) createContainer(address string, name string, comp *componenttree.Component) (container.CreateResponse, error) {
	var deviceMappings []container.DeviceMapping
	info := comp.GetDockerSpec()
	if info == nil {
		return container.CreateResponse{}, fmt.Errorf("No DockerSpec found for component at address %s", address)
	}

	// Port bindings
	for _, port := range info.Ports {
		var mode string = "rwm"
		// TODO: Implement different modes for port mappings if necessary
		// if path.Mode != "" {
		// 	mode = path.Mode
		// }

		deviceMappings = append(deviceMappings, container.DeviceMapping{
			PathOnHost:        port.Source,
			PathInContainer:   port.Target,
			CgroupPermissions: mode,
		})
	}

	// Volume bindings
	var volumeBinds []string
	for _, volume := range info.Volumes {
		var mode string = "rw"
		volumeBinds = append(volumeBinds, fmt.Sprintf("%s:%s:%s", volume.Source, volume.Target, mode))
	}

	// Create container
	resp, err := c.cli.ContainerCreate(c.ctx,
		&container.Config{
			Image: strings.ToLower(name),
			Labels: map[string]string{
				"runtime_hash": c.runtimeHash,
			},
		},
		&container.HostConfig{
			Binds: volumeBinds,
			Resources: container.Resources{
				Devices: deviceMappings,
			},
		},
		nil, nil, name)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// Start a docker container by ID.
func (c *DockerClient) startDockerContainer(ID string) error {
	if err := c.cli.ContainerStart(c.ctx, ID, container.StartOptions{}); err != nil {
		return err
	}
	return nil
}
