.PHONY: build deps run clean

# Variables
BINARY_NAME=helios
BUILD_DIR=bin
MAIN_PATH=./cmd/helios

DOCKER_DISABLED=1 # 1=true, 0=false
export DOCKER_DISABLED

# Commands
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

deps:
	go mod download
	go mod tidy

run:
	go run $(MAIN_PATH)	

clean:
	rm -rf $(BUILD_DIR)
	go clean
