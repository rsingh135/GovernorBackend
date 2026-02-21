#!/bin/bash
# Script to fix Go dependencies and certificate issues

set -e

echo "Attempting to download Go dependencies..."

# Try with default settings first
if go mod tidy; then
    echo "✓ Successfully downloaded dependencies!"
    exit 0
fi

echo "Certificate issue detected. Trying workarounds..."

# Try disabling checksum verification temporarily
export GOSUMDB=off
if go mod tidy; then
    echo "✓ Downloaded dependencies (checksum verification disabled)"
    echo "⚠ Warning: Re-enable checksum verification after fixing certificates:"
    echo "   unset GOSUMDB"
    exit 0
fi

# Try using direct mode
export GOPROXY=direct
export GOSUMDB=off
if go mod tidy; then
    echo "✓ Downloaded dependencies (direct mode, checksum disabled)"
    exit 0
fi

echo "✗ Failed to download dependencies. Please:"
echo "1. Check your internet connection"
echo "2. Fix certificate issues (see FIX_CERTIFICATES.md)"
echo "3. Try running: go mod tidy"
exit 1
