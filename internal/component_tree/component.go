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

func NewComponent(dockerSpec *config.DockerSpec) *Component {
	return &Component{
		dockerSpec: dockerSpec,
	}
}

type DockerConn struct {
	ContainerID string
}

type TransportConn struct {
	Conn net.Conn
}

func GetDockerSpec(c *Component) *config.DockerSpec {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.dockerSpec
}
