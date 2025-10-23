#!/usr/bin/env bash

# Setup git hooks for k8s-agent project

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GIT_DIR="$(git rev-parse --git-dir)"

echo "Setting up git hooks..."

# Make hooks executable
chmod +x "$SCRIPT_DIR/pre-commit"
chmod +x "$SCRIPT_DIR/commit-msg"

# Create symlinks
ln -sf "../../githooks/pre-commit" "$GIT_DIR/hooks/pre-commit"
ln -sf "../../githooks/commit-msg" "$GIT_DIR/hooks/commit-msg"

echo "✓ Git hooks installed successfully!"
echo ""
echo "Installed hooks:"
echo "  - pre-commit: Runs formatting, linting, and basic checks"
echo "  - commit-msg: Validates commit message format (Conventional Commits)"
echo ""
echo "To bypass hooks (not recommended), use: git commit --no-verify"
