# Project variables
BINARY_NAME := "clean-wizard"

# Build binary for current platform
build:
    @echo "🔨 Building {{BINARY_NAME}}..."
    go build -o {{BINARY_NAME}} ./cmd/clean-wizard
    @echo "✅ Build complete: ./{{BINARY_NAME}}"

# Install binary locally
install-local:
    @echo "📦 Installing {{BINARY_NAME}} locally..."
    go install ./cmd/clean-wizard
    @echo "✅ Installation complete"

# Clean build artifacts
clean:
    @echo "🧹 Cleaning build artifacts..."
    rm -f {{BINARY_NAME}}
    go clean

# Run tests
test:
    @echo "🧪 Running tests..."
    go test -v ./...

# Run tests with coverage
test-coverage:
    @echo "🧪 Running tests with coverage..."
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Format code
format:
    @echo "🎨 Formatting code..."
    go fmt ./...
    goimports -w .

# Clean everything (including caches)
clean-all: clean
    @echo "🧹 Cleaning all caches..."
    go clean -modcache
    rm -f coverage.out coverage.html

# Install dependencies
deps:
    @echo "📦 Installing dependencies..."
    go mod download
    go mod tidy

# Run application
run: build
    @echo "🚀 Running {{BINARY_NAME}}..."
    ./{{BINARY_NAME}} --help

# Continuous Integration pipeline
ci: build test
    @echo "✅ CI pipeline completed successfully"

# Default recipe
default:
    @just --list

# Fix module issues
fix-modules:
    @echo "🔧 Fixing module cache..."
    go clean -modcache
    go mod tidy
    go mod download
    go mod verify
    @echo "✅ Modules fixed"
