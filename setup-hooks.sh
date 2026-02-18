#!/bin/bash
# Setup script for configuring Git hooks and security tools

set -e

echo "=== ytdl-go Security Setup ==="
echo ""

# Configure Git to use custom hooks directory
echo "Configuring Git hooks..."
git config core.hooksPath .githooks
echo "✓ Git hooks configured"
echo ""

# Check if gitleaks is installed
echo "Checking for gitleaks..."
if command -v gitleaks &> /dev/null; then
    echo "✓ gitleaks is already installed ($(gitleaks version 2>&1 | head -1))"
else
    echo "⚠ gitleaks is not installed"
    echo ""
    echo "Please install gitleaks to enable pre-commit secret scanning:"
    echo ""
    echo "  macOS:    brew install gitleaks"
    echo "  Linux:    Download from https://github.com/gitleaks/gitleaks/releases"
    echo "  Windows:  Download from https://github.com/gitleaks/gitleaks/releases"
    echo ""
    echo "After installing, run this script again to verify."
fi
echo ""

# Test the hook
echo "Testing pre-commit hook..."
if [ -x .githooks/pre-commit ]; then
    echo "✓ Pre-commit hook is executable"
else
    echo "⚠ Pre-commit hook is not executable, fixing..."
    chmod +x .githooks/pre-commit
    echo "✓ Fixed permissions"
fi
echo ""

echo "=== Setup Complete ==="
echo ""
echo "Your repository is now configured for secret scanning."
echo "All commits will be automatically scanned for secrets."
echo ""
echo "To temporarily skip the pre-commit hook (NOT RECOMMENDED):"
echo "  git commit --no-verify"
echo ""
echo "For more information, see docs/SECRET_MANAGEMENT.md"
