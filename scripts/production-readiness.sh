#!/bin/bash

echo "🧹 Clean Wizard Production Readiness Script"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "OK")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "WARN")
            echo -e "${YELLOW}⚠️  $message${NC}"
            ;;
        "FAIL")
            echo -e "${RED}❌ $message${NC}"
            ;;
    esac
}

# Check prerequisites
echo "📋 Checking prerequisites..."

# Check Go version
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | cut -d' ' -f3)
    print_status "OK" "Go installed: $GO_VERSION"
else
    print_status "FAIL" "Go not installed"
    exit 1
fi

# Check disk space
DISK_USAGE=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
if [ $DISK_USAGE -lt 95 ]; then
    print_status "OK" "Disk space: ${DISK_USAGE}% used"
else
    print_status "WARN" "Disk space: ${DISK_USAGE}% used (getting low)"
fi

# Setup build tools
echo ""
echo "🔧 Setting up build tools..."
if [ -f "scripts/setup-build-tools.sh" ]; then
    ./scripts/setup-build-tools.sh
else
    print_status "FAIL" "Build tools setup script not found"
    exit 1
fi

# Build the project
echo ""
echo "🏗️  Building project..."
if go build -o clean-wizard ./cmd/clean-wizard; then
    print_status "OK" "Build successful"
else
    print_status "FAIL" "Build failed"
    exit 1
fi

# Run tests
echo ""
echo "🧪 Running tests..."
TEST_OUTPUT=$(go test -v ./... 2>&1)
if echo "$TEST_OUTPUT" | grep -q "FAIL"; then
    print_status "WARN" "Some tests failed (details below)"
    echo "$TEST_OUTPUT" | grep -A 5 "FAIL"
else
    print_status "OK" "All tests passed"
fi

# Check if binary works
echo ""
echo "🚀 Testing binary..."
if ./clean-wizard | grep -q "Clean Wizard"; then
    print_status "OK" "Binary is functional"
else
    print_status "FAIL" "Binary not working properly"
fi

# Production readiness checklist
echo ""
echo "📊 Production Readiness Checklist:"

# Check for critical files
CRITICAL_FILES=(
    "README.md"
    "LICENSE"
    "go.mod"
    "cmd/clean-wizard/main.go"
    "Justfile"
)

for file in "${CRITICAL_FILES[@]}"; do
    if [ -f "$file" ]; then
        print_status "OK" "$file exists"
    else
        print_status "FAIL" "$file missing"
    fi
done

# Check if documentation exists
if [ -d "docs" ] && [ "$(ls -A docs)" ]; then
    print_status "OK" "Documentation exists"
else
    print_status "WARN" "Documentation missing or empty"
fi

# Summary
echo ""
echo "📈 Summary:"
echo "  - Build: $(ls -la clean-wizard 2>/dev/null | wc -l) > 0 && echo '✅' || echo '❌'"
echo "  - Tests: $(echo "$TEST_OUTPUT" | grep -c "PASS" || echo 0) passing"
echo "  - Disk: ${DISK_USAGE}% used"

# Next steps
echo ""
echo "🎯 Next Steps:"
echo "1. Address any failing tests above"
echo "2. Review and enhance documentation"
echo "3. Set up CI/CD pipeline"
echo "4. Perform security audit"
echo "5. Create release plan"

echo ""
echo "✨ Clean Wizard is getting closer to production!"