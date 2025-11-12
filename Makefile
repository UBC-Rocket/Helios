.PHONY: build deps run clean

# Variables
BINARY_NAME=helios
BUILD_DIR=bin
MAIN_PATH=./cmd/helios
DOCKER_DISABLED=1

# Commands
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

deps:
	go mod download
	go mod tidy

run:
	export DOCKER_DISABLED
	go run $(MAIN_PATH)	

clean:
	rm -rf $(BUILD_DIR)
	go clean
