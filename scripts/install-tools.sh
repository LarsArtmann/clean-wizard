#!/bin/bash

echo "🔧 Installing missing dependencies..."

# Install goimports for better formatting
go install golang.org/x/tools/cmd/goimports@latest

echo "✅ Dependencies installed"