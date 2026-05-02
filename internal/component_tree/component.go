package component_tree

import (
	"net"
	"sync"

	"helios/generated/config"
	"helios/generated/transport"
)

type Component struct {
	mu            sync.RWMutex
	dockerSpec    *config.DockerSpec
	dockerConn    *DockerConn
	transportConn *TransportConn
	events        map[string]*transport.Event
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
		events:     make(map[string]*transport.Event),
	}
}

func (c *Component) attachConnection(conn net.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.transportConn = &TransportConn{Conn: conn}
}

// The server owns socket lifetime; detach only clears component runtime state.
func (c *Component) detachConnection() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.transportConn = nil
}

func (c *Component) setEvent(eventName string, event *transport.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.events[eventName] = event
}

func (c *Component) getEvent(eventName string) (*transport.Event, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	event, ok := c.events[eventName]
	return event, ok
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

func (c *Component) close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close the transport connection if it exists
	if c.transportConn != nil && c.transportConn.Conn != nil {
		err := c.transportConn.Conn.Close()
		c.transportConn = nil
		return err
	}

	return nil
}
