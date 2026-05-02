package component_tree

import (
	"net"
	"sync"

	configpb "helios/generated/config"
)

type Component struct {
	mu            sync.RWMutex
	dockerSpec    *configpb.DockerSpec
	dockerConn    *DockerConn
	transportConn *TransportConn
}

func NewComponent(dockerSpec *configpb.DockerSpec) *Component {
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
