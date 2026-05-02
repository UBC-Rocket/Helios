package core

import (
	"context"
	"os"
	"fmt"
	"strings"

	configpb "helios/generated/config"
	componenttree "helios/internal/component_tree"
	"helios/internal/logger"
	"helios/internal/docker"

	"google.golang.org/protobuf/encoding/protojson"
)

type Core struct {
	ctx context.Context
	// Unique identifier for the specific Helios instance.
	// Allows re-attaching to the same docker containers if the central helios instance is restarted.
	runtimeHash string
	usingDocker bool
	dockerClient *docker.DockerClient
	// server       *server.Server
	// Represents a tree hierarchy of components.
	tree *componenttree.ComponentTree
}

func Initialize(ctx context.Context, runtimeHash string, usingDocker bool) *Core {
	logger.Infow("Initializing Helios Core", "runtimeHash", runtimeHash, "usingDocker", usingDocker)

	return &Core{
		ctx:         ctx,
		runtimeHash: runtimeHash,
		usingDocker: usingDocker,
		tree:        componenttree.New(),
	}
}

func (c *Core) InitializeComponentTree(path string) error {
	logger.Infow("Initializing component tree", "path", path)

	if strings.TrimSpace(path) == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	logger.Debugw("Reading component tree", "data", string(data))
	cfg := &configpb.ComponentTree{}
	if err := protojson.Unmarshal(data, cfg); err != nil {
		return err
	}

	tree, err := BuildComponentTree(cfg)
	if err != nil {
		return err
	}

	c.tree = tree
	return nil
}

/*
 * Initializes the Docker runtime.
 */
func (c *Core) InitializeDockerRuntime(socketPath string) error {
	if !c.usingDocker {
		return nil
	}
	if c.dockerClient != nil {
		logger.Error("Failed to initialize docker runtime")
		return fmt.Errorf("Docker runtime already initialized")
	}
	logger.Infow("Initializing docker runtime", "socketPath", socketPath, "runtimeHash", c.runtimeHash)
	c.dockerClient = docker.NewDockerClient(c.ctx)
	if err := c.dockerClient.Initialize(); err != nil {
		return err
	}
	if err := c.dockerClient.StartConfigured(); err != nil {
		return err
	}
}

/*
 * Initializes the transport server.
 */
func (c *Core) InitializeServer(address string) error {
	// if c.server != nil {
	// 	return fmt.Errorf("Server already initialized")
	// }
	// logger.Infow("Initializing server", "address", address)
	// srv, err := server.StartServer(c.ctx, address, c)
	// if err != nil {
	// 	return err
	// }
	// c.server = srv
	return nil
}

func (c *Core) Close() error {
	if c.dockerClient != nil {
		err := c.dockerClient.Close()
		if err != nil {
			logger.Warnw("Failed to close docker runtime", "error", err)
		}
	}

	// if c.server != nil {
	// 	err := c.server.Close()
	// 	if err != nil {
	// 		logger.Warnw("Failed to close server", "error", err)
	// 	}
	// }

	// Close component tree
	if c.tree != nil {
		err := c.tree.Close()
		if err != nil {
			logger.Warnw("Failed to close component tree", "error", err)
		}
	}

	return nil
}
