package main

import (
	"os"
	"sync"

	"helios/internal/client"
)

// type DockerResponse struct {
// 	ContainerID string
// 	Error       error
// }

func main() {
	var wg sync.WaitGroup

	runtime_hash := os.Getenv("RUNTIME_HASH")
	docker_disabled := os.Getenv("DOCKER_DISABLED")
	component_tree_path := os.Getenv("COMPONENT_TREE_PATH")

	wg.Add(1)

	runType := "docker"
	if docker_disabled == "1" { runType = "local" }

	cli := client.Initialize(runType, runtime_hash)
	defer cli.Close()

	cli.InitializeComponentTree(component_tree_path)

	// should we call to start components seperately?
	// Initializecomponenttree will automatically start components as of now
	cli.StartAllComponents()



	//! TEMP
	// if docker_disabled == "1" {
	// 	pkt, _ := commhandler.CreateTransportPacket(123, "123", []byte("Test Data"))
	// 	data, _ := commhandler.MarshalTransportPacket(pkt)
	// 	fmt.Println("Marshalled Packet Data:", data)

	// 	fmt.Println("Docker is disabled. Exiting.")
	// 	return // TODO: Implement the local network stuff here instead
	// 	// Probably seperate into seperate function files
	// 	// I.e. one for docker.go, the other for local.go or something
	// }


	// Start all initial containers
	//dh.StartAllContainers(&images, runtime_hash, HeliosNet)

	// Stop the main program from terminating
	// All code is now running in goroutines
	wg.Wait()
}

// !TEMP: Move to commhandler
// Listen for messages from a specific connection
// func listenForMessages(name string, c net.Conn) {
// 	packet := []byte("Hello from Helios!")

// 	c.Write(packet)

// 	//fmt.Printf("Serving %s\n", c.RemoteAddr().String())
// 	p := make([]byte, 0, 4096)
// 	tmp := make([]byte, 4096)
// 	defer c.Close()
// 	for {
// 		n, err := c.Read(tmp)
// 		if err != nil {
// 			if err != io.EOF {
// 				fmt.Println("read error:", err)
// 			}
// 			//println("END OF FILE")
// 			break
// 		}

// 		// If server received some bytes
// 		if n > 0 {
// 			p = append(p[:], tmp[:n]...)
// 		}

// 		end := byte('~')
// 		if p[len(p)-1] == end {
// 			break
// 		}
// 	}
// 	//num, _ := c.Write(p)
// 	fmt.Printf("%s - Received %d bytes, the payload is %s\n", name, len(p), string(p))
// 	listenForMessages(name, c)
// }

//! Move to client
// Handle a new connection passed to the channel
// func handleNewConnection(images []dockerhandler.ImageInfo, conn chan dockerhandler.NewConnection) {
// 	for {
// 		c := <-conn

// 		for _, img := range images {
// 			if img.Name == c.Name {
// 				img.Conn = c.Conn
// 				break
// 			}
// 		}
// 		fmt.Printf("New client connected: %s\n", c.Name)
// 		go listenForMessages(c.Name, c.Conn)
// 	}
// }
