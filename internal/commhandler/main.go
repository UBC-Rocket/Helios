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
	conn        net.Conn
	coreOutChan chan []byte //Outgoing messages from handler to HeliosCore; Write
	coreInChan  chan []byte //Incoming messages from HeliosCore to handler; Read
	contOutChan chan []byte //Outgoing messages from handler to container;  Write //TODO: Is this necessary, or can we just write to port?
	contInChan  chan []byte //Incoming messages from container to handler;  Read
	recentPacket string		 //Placeholder for most recent packet
}

// Create a new communication client for a specific container
func NewCommClient(c net.Conn) *CommClient {
	client := CommClient{
		//ContainerID: containerID,
		conn:        c,
		coreOutChan: make(chan []byte, CHAN_BUFFER_SIZE),
		coreInChan:  make(chan []byte, CHAN_BUFFER_SIZE),
		contOutChan: make(chan []byte, CHAN_BUFFER_SIZE),
		contInChan:  make(chan []byte, CHAN_BUFFER_SIZE),
	}

	go client._listenForContainerMessages()
	go client._handleMessages()

	return &client
}

// Listen for incoming container messages and output to ContInChan
func (c *CommClient) _listenForContainerMessages() {
	tmp := make([]byte, PORT_READ_BUFFER_SIZE)

	for {
		n, _ := c.conn.Read(tmp)

		if n > 0 {
			c.contInChan <- tmp[:n]
		}
	}
}

func (c *CommClient) _handleMessages() {
	select {
	case msg := <-c.coreInChan:
		// Process any incoming stuff from HeliosCore here
		c.conn.Write(msg)
	case msg := <-c.contInChan:
		c.coreOutChan <- msg // If we want to send it to Core
		go c._broadcast(msg)
	case msg := <-c.contOutChan:
		c.conn.Write(msg)
	}
}

func (c *CommClient) _broadcast(msg []byte) {

}
