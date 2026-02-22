package commhandler

import (
	"net"
)

func (c *CommClient) GetConn() net.Conn {
	return c.conn
}

func (c *CommClient) SetConn(conn net.Conn) {
	c.conn = conn
}

func (c *CommClient) GetRecentPacket() string {
	return c.recentPacket
}

func (c *CommClient) GetCoreOutChan() chan []byte {
	return c.coreOutChan
}

func (c *CommClient) GetCoreInChan() chan []byte {
	return c.coreInChan
}

func (c *CommClient) GetContOutChan() chan []byte {
	return c.contOutChan
}

func (c *CommClient) GetContInChan() chan []byte {
	return c.contInChan
}