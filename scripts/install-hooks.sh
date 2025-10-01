#!/bin/bash

# Install git hooks for ZoteroFlow2

set -e

echo "🔧 Installing git hooks..."

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Ensure scripts are executable
chmod +x "$SCRIPT_DIR/pre-commit.sh"
chmod +x "$SCRIPT_DIR/pre-push.sh"
chmod +x "$SCRIPT_DIR/commit-msg.sh"

# Create .git/hooks directory if it doesn't exist
mkdir -p "$PROJECT_ROOT/.git/hooks"

# Install pre-commit hook
echo "📝 Installing pre-commit hook..."
cp "$SCRIPT_DIR/pre-commit.sh" "$PROJECT_ROOT/.git/hooks/pre-commit"

# Install pre-push hook
echo "📝 Installing pre-push hook..."
cp "$SCRIPT_DIR/pre-push.sh" "$PROJECT_ROOT/.git/hooks/pre-push"

# Install commit-msg hook
echo "📝 Installing commit-msg hook..."
cp "$SCRIPT_DIR/commit-msg.sh" "$PROJECT_ROOT/.git/hooks/commit-msg"

echo "✅ Git hooks installed successfully!"
echo ""
echo "📋 Installed hooks:"
echo "   • pre-commit  - Runs formatting, vetting, and basic checks before each commit"
echo "   • pre-push    - Runs full test suite and coverage before each push"
echo "   • commit-msg - Validates commit message format"
echo ""
echo "🚀 Your development workflow is now automated!"