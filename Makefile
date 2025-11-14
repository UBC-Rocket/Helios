.PHONY: build deps run clean proto

# Variables
BINARY_NAME=helios
BUILD_DIR=bin
MAIN_PATH=./cmd/helios
PROTO_SOURCE_DIR=proto
PROTO_LANGS=go python
PROTO_BUILD_DIR=generated

# Find all .proto files in the proto directory and subdirectories
PROTO_SRC := $(wildcard $(PROTO_SOURCE_DIR)/**/*.proto)

# 1=true, 0=false
DOCKER_DISABLED=1
export DOCKER_DISABLED

# Commands
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

	go mod download
	go mod tidy

run:
	go run $(MAIN_PATH)	

clean:
	rm -rf $(BUILD_DIR)
	go clean

proto:
	$(foreach f, $(PROTO_LANGS), \
		$(call build_proto,$(f)) \
	)

define build_proto
	@echo "Generating protobuf files for language: $(1)"
	protoc -I=$(PROTO_SOURCE_DIR) --$(1)_out=$(PROTO_BUILD_DIR)/$(1) $(PROTO_SRC)

endef