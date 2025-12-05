package client

import (
	"math/rand"
	"strconv"
	"testing"
)

func TestInitializeLocalClient(t *testing.T) {
	// Generate a random hash for testing
	hash := strconv.Itoa(rand.Int())

	cli := initializeLocalClient(hash)

	if cli.runtimeHash != hash {
		t.Errorf("Expected runtime_hash to be %s, got %s", hash, cli.runtimeHash)
	}
}

func TestInitializeDockerClient(t *testing.T) {
	// Generate a random hash for testing
	hash := strconv.Itoa(rand.Int())

	cli := initializeDockerClient(hash)

	if cli.runtimeHash != hash {
		t.Errorf("Expected runtimeHash to be %s, got %s", hash, cli.runtimeHash)
	}

	if cli.dc != nil {
		t.Error("Expected Docker client to be nil before initialization")
	}

	if cli.tree == nil {
		t.Error("Expected tree map to be initialized, got nil")
	}
}

// func TestDockerInitialization(t *testing.T) {
// 	// Generate a random hash for testing
// 	hash := strconv.Itoa(rand.Int())

// 	cli := initializeDockerClient(hash)
// 	cli.Initialize()

// 	// Check if Docker client is initialized
// 	if cli.dc == nil {
// 		t.Error("Expected Docker client to be initialized, got nil")
// 	}

// 	if cli.dc.GetContext() == nil {
// 		t.Error("Expected Docker client context to be initialized, got nil")
// 	}

// 	if cli.dc.GetDockerClient() == nil {
// 		t.Error("Expected Docker client instance to be initialized, got nil")
// 	}

// 	if cli.dc.GetNetwork().ID == "" {
// 		t.Error("Expected Docker network to be initialized, got empty ID")
// 	}
// }