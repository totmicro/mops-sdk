.PHONY: build test clean install build-all

PLUGIN_NAME := hello-world-standalone
BINARY_NAME := $(PLUGIN_NAME)

# Default build
build:
	go build -o $(BINARY_NAME) .

# Build for all platforms
build-all:
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -o $(PLUGIN_NAME)-linux-amd64 .
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build -o $(PLUGIN_NAME)-linux-arm64 .
	
	# macOS AMD64 (Intel)
	GOOS=darwin GOARCH=amd64 go build -o $(PLUGIN_NAME)-darwin-amd64 .
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build -o $(PLUGIN_NAME)-darwin-arm64 .
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -o $(PLUGIN_NAME)-windows-amd64.exe .

test:
	go test -v ./...

test-coverage:
	go test -cover ./...

clean:
	rm -f $(PLUGIN_NAME)*
	go clean

install: build
	mkdir -p ~/.mops/plugins
	cp $(BINARY_NAME) ~/.mops/plugins/

lint:
	golangci-lint run

format:
	go fmt ./...

deps:
	go mod tidy
	go mod download

run: build
	./$(BINARY_NAME)

all: format deps test build

help:
	@echo "Available targets:"
	@echo "  build         - Build the plugin binary for current platform"
	@echo "  build-all     - Build for all supported platforms"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install to MOPS plugins directory"
	@echo "  lint          - Run linter"
	@echo "  format        - Format code"
	@echo "  deps          - Download dependencies"
	@echo "  run           - Build and run the plugin"
	@echo "  all           - Format, deps, test, and build"
