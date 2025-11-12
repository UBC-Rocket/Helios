package dockerhandler

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type ImageInfo struct {
	Name string
	Path string
	Tag  string
	ID   string
	Conn net.Conn
}

// TODO: Add component tree here?
type Client struct {
	mu     sync.RWMutex
	cli    *client.Client
	ctx    context.Context
	net    network.CreateResponse
	images []ImageInfo
}

// Initialize the Docker client.
func Initialize() *Client {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	return &Client{
		cli: cli,
		ctx: ctx,
	}
}

// Close the Docker client.
func (c *Client) Close() {
	c.cli.Close()
}

// Get the ID of an existing container given it's name.
// Returns the ID if found and "" if it does not exist.
func (c *Client) GetContainerID(containerName string) (containerID string) {
	list := c.GetContainers()
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
func (c *Client) GetContainers() (summary []container.Summary) {
	list, err := c.cli.ContainerList(c.ctx, container.ListOptions{All: true})
	if err != nil {
		panic(err)
	}
	return list
}

// Create a container using information from the image struct and runtime_hash.
// It should be checked if a container already exists with the same name and hash before calling this function.
func (c *Client) CreateContainer(img ImageInfo, runtime_hash string) (response container.CreateResponse, error error) {
	fmt.Println("Pulling " + img.Tag + " image...")

	// Create container
	resp, err := c.cli.ContainerCreate(c.ctx,
		&container.Config{
			Image: img.Tag,
			Labels: map[string]string{
				"runtime_hash": runtime_hash,
			},
		},
		nil, nil, nil, img.Name)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// Start all containers passed in the images array.
// This function will update the array with the container ID once started
func (c *Client) StartAllContainers(images *[]ImageInfo, runtime_hash string, HeliosNet network.CreateResponse) {
	list := c.GetContainers()

	for _, img := range *images {
		fmt.Println("Checking " + img.Name + " container...")
		var contID string = ""

		for _, x := range list {
			if x.Names[0] == "/"+img.Name {
				existing_hash := x.Labels["runtime_hash"]
				if existing_hash == runtime_hash {

					if x.State == "running" {
						fmt.Println("\n" + img.Name + " container is already running and up-to-date.\n")
						contID = x.ID
						break
					} else if x.State == "exited" {
						fmt.Println("\n" + img.Name + " container is exited but up-to-date. Restarting it...\n")
						if err := c.cli.ContainerStart(c.ctx, x.ID, container.StartOptions{}); err != nil {
							panic(err)
						}
						// Wait until it finishes
						statusCh, errCh := c.cli.ContainerWait(c.ctx, x.ID, container.WaitConditionNotRunning)
						select {
						case err := <-errCh:
							if err != nil {
								panic(err)
							}
						case <-statusCh:
						}

						contID = x.ID
						break
					}
				} else {
					// Remove the container
					fmt.Println(img.Name + " container is outdated. Removing it...")
					if err := c.cli.ContainerRemove(c.ctx, x.ID, container.RemoveOptions{Force: true}); err != nil {
						panic(err)
					}
				}
			}
		}

		if contID == "" {
			fmt.Println("No existing container found. Creating a new one...")
			contResp, contErr := c.CreateContainer(img, runtime_hash)
			if contErr != nil {
				panic(contErr)
			}
			contID = contResp.ID
		}

		fmt.Println("Container ID: " + contID)

		// Save the container ID
		img.ID = contID

		//Add each container to the network before starting them
		fmt.Println("Adding " + img.Name + " container to network...")
		go c.AddContainerToNetwork(contID)
		go c.StartContainer(contID)
	}
}

// Start a docker container by ID.
func (c *Client) StartContainer(ID string) {

	// Start container
	//fmt.Printf("Starting a new %s container...\n", img.Name)
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
