# Cleaning Wizard CLI - Justfile
# This project came from https://github.com/LarsArtmann/Setup-Mac aka https://github.com/LarsArtmann/Setup-Mac/issues/111

# Build variables
BINARY_NAME := "clean-wizard"
VERSION := `git describe --tags --always --dirty`
COMMIT := `git rev-parse --short HEAD`
DATE := `date -u '+%Y-%m-%d %H:%M:%S UTC'`
LDFLAGS := "-ldflags \"-X 'main.version=" + VERSION + "' -X 'main.commit=" + COMMIT + "' -X 'main.date=" + DATE + "'\""

# Default recipe
default:
    @just --list

# Build the binary for current platform
build:
    @echo "🔨 Building {{BINARY_NAME}}..."
    go build -ldflags {{LDFLAGS}} -o {{BINARY_NAME}} ./cmd/clean-wizard
    @echo "✅ Build complete: ./{{BINARY_NAME}}"

# Clean build artifacts
clean:
    @echo "🧹 Cleaning build artifacts..."
    rm -f {{BINARY_NAME}}
    go clean
    @echo "✅ Clean complete"

# Install binary to system
install: build
    @echo "📦 Installing {{BINARY_NAME}}..."
    sudo install {{BINARY_NAME}} /usr/local/bin/
    @echo "✅ Installation complete"

# Run all tests
test:
    @echo "🧪 Running tests..."
    go test -v ./...

# Run tests with coverage
test-coverage:
    @echo "📊 Running tests with coverage..."
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "✅ Coverage report generated: coverage.html"

# Run specific test
test-specific package:
    @echo "🧪 Running tests for {{package}}..."
    go test -v ./{{package}}

# Watch for changes and run tests
test-watch:
    @echo "👀 Watching for changes..."
    find . -name "*.go" | entr -r just test

# Format Go code
fmt:
    @echo "🎨 Formatting Go code..."
    go fmt ./...
    @echo "✅ Code formatted"

# Run go vet
vet:
    @echo "🔍 Running go vet..."
    go vet ./...
    @echo "✅ Vet complete"

# Run linter (requires golangci-lint)
lint:
    @echo "🔍 Running linter..."
    golangci-lint run
    @echo "✅ Lint complete"

# Check all formatting and linting
check: fmt vet lint
    @echo "✅ All checks passed"

# Download dependencies
deps:
    @echo "📦 Downloading dependencies..."
    go mod download
    go mod tidy
    @echo "✅ Dependencies updated"

# Update dependencies
deps-update:
    @echo "🔄 Updating dependencies..."
    go get -u ./...
    go mod tidy
    @echo "✅ Dependencies updated"

# Build for all platforms
build-all:
    @echo "🔨 Building for all platforms..."
    GOOS=darwin GOARCH=amd64 go build -ldflags {{LDFLAGS}} -o {{BINARY_NAME}}-darwin-amd64 ./cmd/clean-wizard
    GOOS=darwin GOARCH=arm64 go build -ldflags {{LDFLAGS}} -o {{BINARY_NAME}}-darwin-arm64 ./cmd/clean-wizard
    GOOS=linux GOARCH=amd64 go build -ldflags {{LDFLAGS}} -o {{BINARY_NAME}}-linux-amd64 ./cmd/clean-wizard
    GOOS=linux GOARCH=arm64 go build -ldflags {{LDFLAGS}} -o {{BINARY_NAME}}-linux-arm64 ./cmd/clean-wizard
    @echo "✅ Cross-platform builds complete"

# Run development server/watch
dev:
    @echo "🚀 Starting development mode..."
    find . -name "*.go" | entr -r just build && ./{{BINARY_NAME}} --help

# Show version
version:
    @echo "📋 {{BINARY_NAME}} version info:"
    @echo "Version: {{VERSION}}"
    @echo "Commit: {{COMMIT}}"
    @echo "Date: {{DATE}}"

# Show project info
info:
    @echo "📋 Project: Cleaning Wizard CLI"
    @echo "Description: A comprehensive CLI/TUI tool for system cleanup"
    @echo "Repository: https://github.com/LarsArtmann/clean-wizard"
    @echo "Origin: https://github.com/LarsArtmann/Setup-Mac/issues/111"

# Create release
release: clean test build-all
    @echo "🚀 Creating release..."
    @echo "Tag: {{VERSION}}"
    @echo "Commit: {{COMMIT}}"
    @echo "Files ready for release"

# Setup development environment
setup:
    @echo "⚙️ Setting up development environment..."
    go mod download
    which just || (echo "📦 Installing just..." && curl -LSsf https://just.systems/install.sh | bash)
    which golangci-lint || (echo "📦 Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
    @echo "✅ Development environment ready"

# Run CI checks
ci: test check
    @echo "✅ All CI checks passed"

# Run integration tests
integration-test:
    @echo "🧪 Running integration tests..."
    go test -v -tags=integration ./...

# Run benchmarks
benchmark:
    @echo "📊 Running benchmarks..."
    go test -bench=. -benchmem ./...

# Generate documentation
docs:
    @echo "📚 Generating documentation..."
    @echo "✅ Documentation generated"

# Docker build
docker-build:
    @echo "🐳 Building Docker image..."
    docker build -t {{BINARY_NAME}}:{{VERSION}} .
    @echo "✅ Docker image built"

# Docker run
docker-run: docker-build
    @echo "🐳 Running Docker container..."
    docker run --rm -it {{BINARY_NAME}}:{{VERSION}}

# Security scan
security:
    @echo "🔒 Running security scan..."
    go list -json -m all | nancy sleuth
    @echo "✅ Security scan complete"

# Profiling
profile:
    @echo "📊 Running with profiling..."
    go build -ldflags {{LDFLAGS}} -o {{BINARY_NAME}}-profile ./cmd/clean-wizard
    ./{{BINARY_NAME}}-profile --cpuprofile=cpu.prof --memprofile=mem.prof scan
    go tool pprof cpu.prof

# Clean everything (including caches)
clean-all: clean
    @echo "🧹 Cleaning all caches..."
    go clean -cache
    go clean -modcache
    @echo "✅ Deep clean complete"