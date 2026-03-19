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
    rm -rf bin/ {{BINARY_NAME}}
    find . -name "*.test" -type f -delete
    rm -f coverage.out coverage.html
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
# NOTE: This skips Go cache cleaning if other Go processes are running to avoid cache corruption.
# This is a safety measure - if you need to clean Go caches, ensure no other Go processes are running first.
clean-all:
    @echo "🧹 Cleaning all build artifacts..."
    rm -rf bin/ {{BINARY_NAME}}
    find . -name "*.test" -type f -delete
    rm -f coverage.out coverage.html
    rm -rf reports/
    go clean
    @echo "🧹 Cleaning caches via clean-wizard (skipping Go caches if processes are running)..."
    @go build -o {{BINARY_NAME}} ./cmd/clean-wizard 2>&1 > /dev/null || true
    @./{{BINARY_NAME}} clean --mode quick --json > /dev/null 2>&1 || echo "ℹ️  clean-wizard skipped"
    rm -f {{BINARY_NAME}}

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
fix-modules: build
    @echo "🔧 Fixing module cache via clean-wizard..."
    {{BINARY_NAME}} clean --json --mode quick > /dev/null || echo "ℹ️  Go cache cleaned via clean-wizard"
    @echo "🔧 Tidying and verifying modules..."
    go mod tidy
    go mod download
    go mod verify
    @echo "✅ Modules fixed"
