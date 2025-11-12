package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"helios/internal/dockerhandler"
)

const (
	SELF_NAME = "HeliosCore"
	NET_NAME = "HeliosNet"
	PORT = "5000"
)

type DockerResponse struct {
	ContainerID string
	Error       error
}

var wg sync.WaitGroup


func main() {
	fmt.Println("Main function started")

	images := []dockerhandler.ImageInfo{
		{"RocketDecoder", "../RocketDecoder", "rocketdecoder:latest", "", nil},
		//{"UI", "../UI", "ui:latest", "", nil},
		{"Livestream", "../Livestream", "livestream:latest", "", nil},
	}

	// Create new docker client
	dh := dockerhandler.Initialize()
	runtime_hash := os.Getenv("RUNTIME_HASH")

	defer dh.Close()

	// Start the docker network
	HeliosNet, netErr := dh.StartDockerNetwork(NET_NAME)
	if netErr != nil { panic(netErr) }

	// Start the port connection and pass a channel for new connection updates
	wg.Add(1)
	connectionChan := make(chan dockerhandler.NewConnection)
	go dh.StartPortConnection(PORT, connectionChan)
	go handleNewConnection(images, connectionChan)

	// Get own container's ID
	self_id := dh.GetContainerID(SELF_NAME)
	if self_id == "" { panic("HeliosCore ID not found.")}

	// Add HeliosCore to the network
	fmt.Println("Adding HeliosCore to HeliosNet...")
	dh.AddContainerToNetwork(self_id)

	// Start all initial containers
	dh.StartAllContainers(&images, runtime_hash, HeliosNet)

	// Stop the main program from terminating
	// All code is now running in goroutines
	wg.Wait()
}

// Listen for messages from a specific connection
func listenForMessages(name string, c net.Conn) {
  packet := []byte("Hello from HeliosCore!")

	c.Write(packet)

	//fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	p := make([]byte, 0, 4096)
	tmp := make([]byte, 4096)
	defer c.Close()
	for {
		n, err := c.Read(tmp)
		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			//println("END OF FILE")
			break
		}

		// If server received some bytes
		if n > 0 { 
			p = append(p[:], tmp[:n]...)
		}

		end := byte('~')
		if (p[len(p)-1] == end) { break }
	}
	//num, _ := c.Write(p)
	fmt.Printf("%s - Received %d bytes, the payload is %s\n", name, len(p), string(p))
	listenForMessages(name, c)
}

// Handle a new connection passed to the channel
func handleNewConnection(images []dockerhandler.ImageInfo, conn chan dockerhandler.NewConnection) {
	for {
		c := <- conn

		for _, img := range images {
			if img.Name == c.Name {
				img.Conn = c.Conn
				break
			}
		}
		fmt.Printf("New client connected: %s\n", c.Name)
		go listenForMessages(c.Name, c.Conn)
	}
}