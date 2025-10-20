#!/bin/bash

# Setup script for migrate-v2 tool
# This script initializes and updates the git submodules

set -e

echo "Setting up migrate-v2 submodules..."

# Navigate to the repository root
cd "$(git rev-parse --show-toplevel)"

# Initialize submodules if not already initialized
echo "Initializing submodules..."
git submodule init

# Update submodules to their respective branches
echo "Updating provider-v4 to 'v4' branch..."
git submodule update --init --recursive cmd/migrate-v2/external/provider-v4

echo "Updating provider-v5 to 'next' branch..."
git submodule update --init --recursive cmd/migrate-v2/external/provider-v5

# Ensure submodules are on the correct branches
echo "Checking out correct branches..."
(cd cmd/migrate-v2/external/provider-v4 && git checkout v4)
(cd cmd/migrate-v2/external/provider-v5 && git checkout next)

echo "✅ Submodules setup complete!"
echo ""
echo "Provider versions:"
echo -n "  provider-v4: "
(cd cmd/migrate-v2/external/provider-v4 && git rev-parse --abbrev-ref HEAD)
echo -n "  provider-v5: "
(cd cmd/migrate-v2/external/provider-v5 && git rev-parse --abbrev-ref HEAD)
echo ""
echo "You can now build the migrate-v2 tool:"
echo "  cd cmd/migrate-v2 && make build"