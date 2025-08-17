# Makefile for the LLM Gateway

# Define the binary name and path relative to this Makefile
BINARY_NAME=bin/gateway
CMD_DIR=./cmd/gateway
BINARY_PATH=$(CMD_DIR)/$(BINARY_NAME)
SOURCE_PATH=$(CMD_DIR)/main.go

# Default target
all: build

# Build the Go application
build:
	@echo "Building the application..."
	@go build -o $(BINARY_PATH) $(SOURCE_PATH)
	@echo "Build complete. Binary is at $(BINARY_PATH)"

# Run the application from the Makefile's directory to ensure it finds config.yaml
run: build
	@echo "Starting the application..."
	@$(BINARY_PATH)

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean the build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_PATH)
	@echo "Cleanup complete."

.PHONY: all build run clean test
