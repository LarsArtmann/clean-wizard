.PHONY: all build install clean test lint fmt vet help

# Build variables
BINARY_NAME=clean-wizard
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT?=$(shell git rev-parse --short HEAD)
DATE?=$(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
LDFLAGS=-ldflags "-X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'"

# Default target
all: clean build

# Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/clean-wizard/

# Install the binary
install: build
	go install $(LDFLAGS) ./cmd/clean-wizard/

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)

# Run tests
test:
	go test -v ./...

# Run linting
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Build and run tests
check: fmt vet test

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/clean-wizard/
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 ./cmd/clean-wizard/
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd/clean-wizard/
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 ./cmd/clean-wizard/

# Show help
help:
	@echo "Available targets:"
	@echo "  all         - Clean and build"
	@echo "  build       - Build the binary"
	@echo "  install     - Install the binary"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linting"
	@echo "  fmt         - Format code"
	@echo "  vet         - Run go vet"
	@echo "  check       - Format, vet, and test"
	@echo "  build-all   - Build for multiple platforms"
	@echo "  help        - Show this help"