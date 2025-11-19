# Project variables
BINARY_NAME := "clean-wizard"

# Build binary for current platform
build:
    @echo "🔨 Building {{BINARY_NAME}}..."
    go build -o {{BINARY_NAME}} ./cmd/clean-wizard
    @echo "✅ Build complete: ./{{BINARY_NAME}}"

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
    @command -v goimports >/dev/null 2>&1 && goimports -w . || echo "⚠️  goimports not available, run 'just install-tools'"

# Install development tools
install-tools:
    @echo "🔧 Installing development tools..."
    go install golang.org/x/tools/cmd/goimports@latest
    @echo "✅ Tools installed"

# Run linter (basic go vet + custom checks)
lint:
    @echo "🔍 Running linting..."
    go vet ./...
    goimports -l .
    @echo "✅ Linting complete"

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
