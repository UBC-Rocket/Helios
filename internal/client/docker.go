package client

import (
	"helios/internal/dockerhandler"
)

const (
	SELF_NAME = "Helios"
	NET_NAME  = "HeliosNet"
	PORT      = "5000"
)

type DockerInterface struct {
	dc *dockerhandler.DockerClient
	runtime_hash string
	tree string // placeholder for local tree object; includes connection object, cont. id, etc
}

func initializeDockerClient(runtime_hash string) *DockerInterface {
	return &DockerInterface{
		dc: &dockerhandler.DockerClient{},
		runtime_hash: runtime_hash,
	}
}

func (x *DockerInterface) Initialize() {
	// Create a new docker client
	x.dc.Initialize()

	// Start the docker network
	_, netErr := x.dc.StartDockerNetwork(NET_NAME)
	if netErr != nil {
		panic(netErr)
	}

	connectionChan := make(chan dockerhandler.NewConnection)
	go x.dc.StartPortConnection(PORT, connectionChan)
	go handleNewConnection(images, connectionChan)

	// Get own container's ID
	self_id := x.dc.GetContainerID(SELF_NAME)
	if self_id == "" {
		panic(SELF_NAME + "ID not found.")
	}

	// Add Helios to the network
	//fmt.Println("Adding", SELF_NAME, "to", NET_NAME, "...")
	x.dc.AddContainerToNetwork(self_id)
}

func (x *DockerInterface) InitializeComponentTree(path string) {
	//tree := extractComponentTree(path)
	return
	// Read the JSON, generate a local object for component tree
}

func (x *DockerInterface) StartComponent(name string) {
	// Start an individual component
	// Then add to network
}

func (x *DockerInterface) StartAllComponents() {
	//Check if the component tree has been inititalized yet before running
}

func (x *DockerInterface) StopComponent(name string) {
  // Stop an individual component
	// docker stop
	// Send a message over to component to get it to stop
}

func (x *DockerInterface) KillComponent(name string) {
  // Kill? an individual component (Stop, remove)
	// docker kill
	// Force kill
}

func (x *DockerInterface) Clean() {
	// Clean component (s)?
}

func (x *DockerInterface) Close() {
	x.dc.Close()
}