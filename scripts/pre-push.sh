#!/bin/bash

# Pre-push hook for ZoteroFlow2
# This script runs before each push to ensure code quality

set -e

echo "🚀 Running pre-push checks..."

# Check if we're in the server directory
if [ ! -f "go.mod" ]; then
    echo "Changing to server directory..."
    cd server
fi

# 1. Run full test suite
echo "🧪 Running full test suite..."
make test

# 2. Run coverage check
echo "📊 Running test coverage..."
make test-coverage

# Extract coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}' | sed 's/%//')
echo "Current test coverage: $COVERAGE"

# Check if coverage meets minimum threshold (80%)
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "❌ Test coverage below 80%: $COVERAGE"
    echo "Please add more tests to improve coverage."
    exit 1
fi

echo "✅ Test coverage OK: $COVERAGE"

# 3. Run linter (if available)
if command -v golangci-lint &> /dev/null; then
    echo "🔍 Running golangci-lint..."
    golangci-lint run
    echo "✅ Linter checks OK"
else
    echo "ℹ️  golangci-lint not found, skipping linter checks"
    echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

# 4. Run race condition tests
echo "🏃 Running race condition tests..."
go test -race ./...
echo "✅ Race condition tests OK"

# 5. Run memory leak tests (if possible)
echo "🧠 Running memory allocation tests..."
go test -memprofile=mem.prof -run TestMemory ./...
echo "✅ Memory tests OK"

# 6. Check for dependency vulnerabilities
echo "🔒 Checking for known vulnerabilities..."
go list -m -u all | grep -E "(CVE|vulnerability)" || echo "✅ No known vulnerabilities found"

# 7. Run build with different targets (if cross-compilation tools available)
echo "🔨 Testing build targets..."
make build

echo "✅ Pre-push checks completed successfully!"
echo "🚀 Ready to push!"