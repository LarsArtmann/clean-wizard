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
    go install golang.org/x/tools/cmd/stringer@latest
    @echo "✅ Tools installed"

# Generate code
generate:
    @echo "🔄 Generating code..."
    go generate ./...
    @echo "✅ Code generation complete"

# Full setup including code generation
setup: install-tools generate
    @echo "✅ Full setup complete"

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
    rm -rf .coverage .reports

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
ci: setup build test
    @echo "✅ CI pipeline completed successfully"

# Production readiness check
prod-ready:
    @echo "🚀 Running production readiness check..."
    @if [ -f "scripts/production-readiness.sh" ]; then \
		./scripts/production-readiness.sh; \
	else \
		echo "❌ Production readiness script not found"; \
	fi

# Default recipe
default:
    @just --list

# Find code duplications in project
find-duplicates threshold="15":
    @echo "\033[1m🔍 FINDING CODE DUPLICATIONS\033[0m"
    @mkdir -p .reports
    @echo "\033[0;36mAnalyzing Go code duplications (threshold: {{threshold}} tokens)...\033[0m"
    @if command -v dupl >/dev/null 2>&1; then \
        echo "\033[0;33m📋 Go Code Duplication Report (dupl)\033[0m"; \
        dupl -t {{threshold}} -v . > .reports/go-duplications.txt 2>&1 || true; \
        dupl -t {{threshold}} -html . > .reports/go-duplications.html 2>&1 || true; \
        echo "  → .reports/go-duplications.txt"; \
        echo "  → .reports/go-duplications.html"; \
        echo ""; \
        echo "\033[0;33m📊 Summary:\033[0m"; \
        DUPL_COUNT=`dupl -t {{threshold}} . 2>/dev/null | grep -c "found" || echo "0"`; \
        echo "  Go duplications found: $$DUPL_COUNT"; \
    else \
        echo "\033[0;31m❌ dupl not found. Installing...\033[0m"; \
        go install github.com/mibk/dupl@latest; \
        dupl -t {{threshold}} -v . > .reports/go-duplications.txt 2>&1 || true; \
        dupl -t {{threshold}} -html . > .reports/go-duplications.html 2>&1 || true; \
    fi
    @echo "\033[0;36mAnalyzing multi-language duplications (jscpd)...\033[0m"
    @if command -v jscpd >/dev/null 2>&1; then \
        echo "\033[0;33m📋 Multi-Language Duplication Report (jscpd)\033[0m"; \
        jscpd . --min-tokens {{threshold}} --reporters json,html --output .reports || true; \
        if [ -f ".reports/jscpd/jscpd-report.json" ]; then \
            echo "  → .reports/jscpd/jscpd-report.json"; \
            echo "  → .reports/jscpd/jscpd-report.html"; \
        fi; \
    else \
        echo "\033[0;33m⚠️  jscpd not found, skipping multi-language analysis.\033[0m"; \
        echo "\033[0;36mTo install: bun install -g jscpd\033[0m"; \
    fi
    @echo ""
    @echo "\033[0;32m✅ Duplication analysis complete!\033[0m"
    @echo "\033[0;36mOpen .reports/go-duplications.html in browser for detailed Go analysis\033[0m"

# Alias for find-duplicates
fd threshold="15": (find-duplicates threshold)
