package component_tree

import (
	"net"
	"sync"

	"helios/generated/config"
)

type Component struct {
	mu            sync.RWMutex
	dockerSpec    *config.DockerSpec
	dockerConn    *DockerConn
	transportConn *TransportConn
}

type DockerConn struct {
	ContainerID string
}

type TransportConn struct {
	Conn net.Conn
}

func NewComponent(dockerSpec *config.DockerSpec) *Component {
	return &Component{
		dockerSpec: dockerSpec,
	}
}

func (c *Component) GetDockerSpec() *config.DockerSpec {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.dockerSpec
}

func (c *Component) SetDockerConnID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// TODO: Check if DockerConn already exists and handle accordingly
	c.dockerConn = &DockerConn{ContainerID: id}
}