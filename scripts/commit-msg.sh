#!/bin/bash

# Simplified commit message hook

set -e

echo "📝 Validating commit message..."

# Get the commit message
COMMIT_MSG_FILE="$1"
if [ ! -f "$COMMIT_MSG_FILE" ]; then
    echo "❌ Commit message file not found"
    exit 1
fi

COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")

# Check if commit message is empty
if [ -z "$COMMIT_MSG" ]; then
    echo "❌ Commit message cannot be empty."
    exit 1
fi

# Get first line
FIRST_LINE=$(echo "$COMMIT_MSG" | head -n1)

# Check length (max 72 characters)
if [ ${#FIRST_LINE} -gt 72 ]; then
    echo "❌ First line too long (max 72 characters): ${#FIRST_LINE}"
    echo "Current: $FIRST_LINE"
    exit 1
fi

# Check for trailing whitespace
if echo "$COMMIT_MSG" | grep -q '[[:space:]]$'; then
    echo "❌ Commit message has trailing whitespace."
    exit 1
fi

# Check for conventional commit format (simplified)
if [[ ! "$FIRST_LINE" =~ ^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\(.+\))?:\ .+ ]]; then
    echo "❌ Commit message should follow format: <type>(<scope>): <description>"
    echo ""
    echo "Valid types: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert"
    echo ""
    echo "Examples:"
    echo "   feat(parser): add MinerU PDF parsing support"
    echo "   fix(core): resolve CSV record encoding issue"
    echo "   docs(readme): update installation instructions"
    echo "   test(parser): add unit tests for PDF parsing"
    echo "   chore(deps): update dependencies"
    exit 1
fi

# Check that first letter is lowercase
if [[ ! "$FIRST_LINE" =~ ^[a-z] ]]; then
    echo "❌ Commit message should start with lowercase letter."
    exit 1
fi

# Check that it doesn't end with period
if [[ "$FIRST_LINE" =~ \.$ ]]; then
    echo "❌ Commit message should not end with a period."
    exit 1
fi

echo "✅ Commit message validation passed!"
echo "📝 Commit message: $FIRST_LINE"