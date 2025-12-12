#!/bin/bash
echo "🔧 Installing missing build tools..."

# Check and install stringer if needed
if ! command -v stringer &> /dev/null; then
    echo "Installing stringer tool..."
    go install golang.org/x/tools/cmd/stringer@latest
    echo "✅ stringer installed"
else
    echo "✅ stringer already installed"
fi

# Generate all required code
echo "🔄 Generating code..."
cd "$(dirname "$0")/.."

# Generate stringers for enums
echo "Generating enum stringers..."
go generate ./...

# Generate any other code
echo "Running additional code generation..."
if [ -f "scripts/generate.sh" ]; then
    ./scripts/generate.sh
fi

echo "✅ Code generation complete"