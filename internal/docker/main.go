package docker

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	ct "helios/internal/component_tree"

	configpb "helios/generated/config"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

const (
	NET_NAME = "HeliosNet"
)

type DockerClient struct {
	cli         *client.Client
	ctx         context.Context
	net         network.CreateResponse
	socketPath  string
	runtimeHash string
}

func NewDockerClient(ctx context.Context, socketPath string, runtimeHash string) *DockerClient {
	return &DockerClient{
		ctx:         ctx,
		socketPath:  socketPath,
		runtimeHash: runtimeHash,
	}
}

// Initialize the Docker client.
func (c *DockerClient) Initialize() error {
	host := "unix://" + c.socketPath // TODO: Support windows??
	cli, err := client.NewClientWithOpts(client.WithHost(host), client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	c.cli = cli
	return nil
}

func (c *DockerClient) StartConfigured(tree *ct.ComponentTree) error {
	if tree == nil {
		return nil
	}

	_, netErr := c.StartDockerNetwork(NET_NAME)
	if netErr != nil {
		return netErr
	}

	return tree.ForEachDockerSpec(func(address string, spec *configpb.DockerSpec) error {
		name := ContainerName(address, spec)

		id := c.StartContainer(name, spec, c.runtimeHash)

		tree.SetDockerContainerID(address, id)
		return nil
	})
}

func (c *DockerClient) SocketPath() string {
	return c.socketPath
}

// Close the Docker client.
func (c *DockerClient) Close() {
	c.cli.Close()
}

// Get the ID of an existing container given it's name.
// Returns the ID if found and "" if it does not exist.
func (c *DockerClient) GetContainerID(containerName string) (containerID string) {
	list := c.GetContainers()
	var contID string = ""

	for _, cont := range list {
		if len(cont.Names) > 0 && cont.Names[0] == "/"+containerName {
			contID = cont.ID
			break
		}
	}

	return contID
}

// Get a list of all containers.
func (c *DockerClient) GetContainers() (summary []container.Summary) {
	list, err := c.cli.ContainerList(c.ctx, container.ListOptions{All: true})
	if err != nil {
		panic(err)
	}
	return list
}

// Create a container using information from the image struct and runtime_hash.
// It should be checked if a container already exists with the same name and hash before calling this function.
func (c *DockerClient) createContainer(name string, spec *configpb.DockerSpec, hash string) (response container.CreateResponse, error error) {
	// Port bindings
	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}
	for _, port := range spec.GetPorts() {
		target := strings.TrimSpace(port.GetTarget())
		if target == "" {
			continue
		}
		protocol := port.GetType()
		if protocol == "" {
			protocol = "tcp"
		}
		dockerPort := nat.Port(target + "/" + protocol)
		exposedPorts[dockerPort] = struct{}{}
		if source := strings.TrimSpace(port.GetSource()); source != "" {
			portBindings[dockerPort] = append(portBindings[dockerPort], nat.PortBinding{HostPort: source})
		}
	}

	// Volume bindings
	var volumeBinds []string
	for _, volume := range spec.GetVolumes() {
		mode := volume.GetMode()
		if mode == "" {
			mode = "rw"
		}
		volumeBinds = append(volumeBinds, fmt.Sprintf("%s:%s:%s", volume.Source, volume.Target, mode))
	}

	// Create container
	resp, err := c.cli.ContainerCreate(c.ctx,
		&container.Config{
			Image:        imageName(name, spec),
			ExposedPorts: exposedPorts,
			Labels: map[string]string{
				"runtime_hash": hash,
			},
		},
		&container.HostConfig{
			Binds:        volumeBinds,
			PortBindings: portBindings,
		},
		nil, nil, name)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// Start a container given its name, tag, and runtime_hash.
// If the container already exists, it will be restarted or removed and recreated if the hash does not match.
// Returns the container ID of the started container.
// Created container will be added to the docker network and started.
func (c *DockerClient) StartContainer(name string, spec *configpb.DockerSpec, hash string) (ID string) {
	list := c.GetContainers()
	var cont container.Summary = container.Summary{}

	// Find container by name to see if it exists
	for _, x := range list {
		if len(x.Names) > 0 && x.Names[0] == "/"+name {
			cont = x
			break
		}
	}

	// If the runtime hash does not match
	if (cont.ID != "") && (cont.Labels["runtime_hash"] != hash) {
		// Remove the container
		if err := c.cli.ContainerRemove(c.ctx, cont.ID, container.RemoveOptions{Force: true}); err != nil {
			fmt.Println("Error removing outdated container:", err)
		}

		// Reset cont to indicate it does not exist
		cont = container.Summary{}

		// If the runtime hash matches, check if is exited or running
	} else if cont.ID != "" {
		if cont.State == "running" {
			// Do nothing
		} else if cont.State == "exited" {
			// Restart it
			go c.startDockerContainer(cont.ID)
		}
	}

	// If container does not exist, create it
	if cont.ID == "" {
		contResp, contErr := c.createContainer(name, spec, hash)
		if contErr != nil {
			// TODO: Handle error properly
			fmt.Println("Error creating container:", contErr)
		}
		cont.ID = contResp.ID
	}

	// Add to network and start container
	go c.AddContainerToNetwork(cont.ID)
	go c.startDockerContainer(cont.ID)
	return cont.ID
}

// Start a docker container by ID.
func (c *DockerClient) startDockerContainer(ID string) {

	// Start container
	if err := c.cli.ContainerStart(c.ctx, ID, container.StartOptions{}); err != nil {
		panic(err)
	}

	// Wait until it finishes
	waitStatusCh, waitErrCh := c.cli.ContainerWait(c.ctx, ID, container.WaitConditionNotRunning)
	select {
	case err := <-waitErrCh:
		if err != nil {
			panic(err)
		}
	case <-waitStatusCh:
	}

	// Get logs
	//TODO: Move this to a seperate logging driver later
	out, err := c.cli.ContainerLogs(c.ctx, ID, container.LogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

func ContainerName(address string, spec *configpb.DockerSpec) string {
	if strings.TrimSpace(spec.GetContainerName()) != "" {
		return spec.GetContainerName()
	}
	return sanitizeContainerName(address)
}

func imageName(name string, spec *configpb.DockerSpec) string {
	image := strings.TrimSpace(spec.GetImage())
	if image == "" {
		image = name
	}
	if strings.TrimSpace(spec.GetTag()) != "" && !strings.Contains(image, ":") {
		image += ":" + spec.GetTag()
	}
	return strings.ToLower(image)
}

func sanitizeContainerName(address string) string {
	sanitized := strings.ToLower(address)
	sanitized = regexp.MustCompile(`[^a-z0-9_.-]+`).ReplaceAllString(sanitized, "-")
	sanitized = strings.Trim(sanitized, "-_.")
	if sanitized == "" {
		return "helios-component"
	}
	return sanitized
}
