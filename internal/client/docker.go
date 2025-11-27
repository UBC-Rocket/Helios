package client

import (
	"sync"

	"helios/generated/go/config"
	"helios/internal/commhandler"
	"helios/internal/dockerhandler"
)

const (
	SELF_NAME = "Helios"
	NET_NAME  = "HeliosNet"
	PORT      = "5000"
)

type ComponentObject struct {
	mu           sync.RWMutex
	containerID  string // Docker ID
	componentID  string
	group        string // Tree group
	path         string
	tag          string // TODO: Do we want to keep this?
	commHandler  commhandler.CommClient
	recentPacket string //Placeholder for most recent packet
}

type DockerInterface struct {
	dc          *dockerhandler.DockerClient
	runtimeHash string
	tree        map[string]*ComponentObject // placeholder for local tree object; includes connection object, cont. id, etc
}

func initializeDockerClient(hash string) *DockerInterface {
	return &DockerInterface{
		dc:          &dockerhandler.DockerClient{},
		runtimeHash: hash,
		tree:        make(map[string]*ComponentObject),
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
	go x.handleNewConnection(connectionChan)

	// Get own container's ID
	self_id := x.dc.GetContainerID(SELF_NAME)
	if self_id == "" {
		panic(SELF_NAME + "ID not found.")
	}

	// Add self to the network
	x.dc.AddContainerToNetwork(self_id)
}

func (x *DockerInterface) InitializeComponentTree(path string) {
	tree := extractComponentTree(path)

	// Recursively add children to tree
	x.addTreeNodes(tree.Root, SELF_NAME)
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

// Recursively checks all branches and adds children to the tree object
func (x *DockerInterface) addTreeNodes(n *config.BaseComponent, g string) {
	switch v := n.NodeType.(type) {
	case *config.BaseComponent_Branch:
		for _, c := range v.Branch.Children {
			x.addTreeNodes(c, g + "_" + n.Name)
		}
	case *config.BaseComponent_Leaf:
		x.tree[n.Name] = &ComponentObject{
			group:       g,
			path:        v.Leaf.Path,
			tag:         v.Leaf.Tag,
			componentID: v.Leaf.Id,
		}
	}
}

// Handles new connections made by Docker through an update channel.
// If the component does not existing in the tree by name, a new one is made.
// A new communications handler is spawned with the connection object and saved to the component.
func (x *DockerInterface) handleNewConnection(conn chan dockerhandler.NewConnection) {
	for {
		c := <-conn

		// Check if component exists in tree
		comp, ok := x.tree[c.Name]

		if ok {
			comp.mu.Lock()

			// Check if an empty communications handler was initialized
			if (commhandler.CommClient{}) == comp.commHandler {
				comp.commHandler = *commhandler.NewCommClient(c.Conn)
			} else {
				comp.commHandler.Conn = c.Conn
			}

			comp.mu.Unlock()
		} else {
			x.tree[c.Name] = &ComponentObject{
				commHandler: *commhandler.NewCommClient(c.Conn),
			}
		}
	}
}
