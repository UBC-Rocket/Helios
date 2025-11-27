package commhandler

import (
	"net"
)

const (
	CHAN_BUFFER_SIZE      = 16384 //16 KB
	PORT_READ_BUFFER_SIZE = 4096  //4 KB
)

type CommClient struct {
	//ContainerID string
	Conn        net.Conn
	CoreOutChan chan []byte //Outgoing messages from handler to HeliosCore; Write
	CoreInChan  chan []byte //Incoming messages from HeliosCore to handler; Read
	ContOutChan chan []byte //Outgoing messages from handler to container;  Write //TODO: Is this necessary, or can we just write to port?
	ContInChan  chan []byte //Incoming messages from container to handler;  Read
}

// Create a new communication client for a specific container
func NewCommClient(c net.Conn) *CommClient {
	client := CommClient{
		//ContainerID: containerID,
		Conn:        c,
		CoreOutChan: make(chan []byte, CHAN_BUFFER_SIZE),
		CoreInChan:  make(chan []byte, CHAN_BUFFER_SIZE),
		ContOutChan: make(chan []byte, CHAN_BUFFER_SIZE),
		ContInChan:  make(chan []byte, CHAN_BUFFER_SIZE),
	}

	go client._listenForContainerMessages()
	go client._handleMessages()

	return &client
}

// Listen for incoming container messages and output to ContInChan
func (c *CommClient) _listenForContainerMessages() {
	tmp := make([]byte, PORT_READ_BUFFER_SIZE)

	for {
		n, _ := c.Conn.Read(tmp)

		if n > 0 {
			c.ContInChan <- tmp[:n]
		}
	}
}

func (c *CommClient) _handleMessages() {
	select {
	case msg := <-c.CoreInChan:
		// Process any incoming stuff from HeliosCore here
		c.Conn.Write(msg)
	case msg := <-c.ContInChan:
		c.CoreOutChan <- msg // If we want to send it to Core
		go c._broadcast(msg)
	case msg := <-c.ContOutChan:
		c.Conn.Write(msg)
	}
}

func (c *CommClient) _broadcast(msg []byte) {

}
