#!/bin/bash

# Pre-commit hook for ZoteroFlow2
# This script runs before each commit to ensure code quality

set -e

echo "🔍 Running pre-commit checks..."

# Check if we're in the server directory
if [ ! -f "go.mod" ]; then
    echo "Changing to server directory..."
    cd server
fi

# 1. Run go fmt
echo "📝 Checking code formatting..."
if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
    echo "❌ Code formatting issues found:"
    gofmt -s -l .
    echo "Please run 'make fmt' to fix formatting issues."
    exit 1
fi
echo "✅ Code formatting OK"

# 2. Run go vet
echo "🔎 Running go vet..."
go vet ./...
echo "✅ Go vet OK"

# 3. Run go mod tidy
echo "🧹 Running go mod tidy..."
go mod tidy
echo "✅ Go mod tidy OK"

# 4. Run tests if test files exist
if ls *_test.go 1> /dev/null 2>&1; then
    echo "🧪 Running tests..."
    go test -v ./...
    echo "✅ Tests OK"
else
    echo "ℹ️  No test files found, skipping tests"
fi

# 5. Check for common Go anti-patterns
echo "🚫 Checking for common issues..."

# Check for TODO/FIXME comments
if grep -r "TODO\|FIXME\|XXX\|HACK" . --include="*.go" | grep -v "vendor" | grep -q .; then
    echo "⚠️  Found TODO/FIXME comments:"
    grep -r "TODO\|FIXME\|XXX\|HACK" . --include="*.go" | grep -v "vendor"
    echo "Please address these items before committing."
fi

# Check for hardcoded credentials
if grep -r "password\|secret\|token\|key.*=" . --include="*.go" --include="*.yaml" --include="*.yml" | grep -v "example\|sample\|test" | grep -q .; then
    echo "⚠️  Found potential hardcoded credentials:"
    grep -r "password\|secret\|token\|key.*=" . --include="*.go" --include="*.yaml" --include="*.yml" | grep -v "example\|sample\|test"
    echo "Please ensure no credentials are committed."
fi

# 6. Check file sizes
echo "📊 Checking file sizes..."
LARGE_FILES=$(find . -type f -name "*.go" -size +100k | head -5)
if [ -n "$LARGE_FILES" ]; then
    echo "⚠️  Found large Go files (>100KB):"
    echo "$LARGE_FILES"
    echo "Consider breaking them into smaller files."
fi

echo "✅ Pre-commit checks completed successfully!"
echo "🚀 Ready to commit!"