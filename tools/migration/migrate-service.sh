#!/bin/bash

# Migration script for restructuring k8s-agent project
# Usage: ./migrate-service.sh <service-name>

set -e

SERVICE=$1
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

if [ -z "$SERVICE" ]; then
    echo "Usage: $0 <service-name>"
    echo "Available services:"
    echo "  - agent-manager"
    echo "  - orchestrator-service"
    echo "  - reasoning-service-go"
    echo "  - auth-service"
    echo "  - gateway-service"
    echo "  - monitor-service"
    echo "  - cluster-service"
    echo "  - collect-agent"
    exit 1
fi

# Normalize service names
case "$SERVICE" in
    agent-manager|orchestrator-service|reasoning-service-go|auth-service|gateway-service|monitor-service|cluster-service|collect-agent)
        SOURCE_DIR="${ROOT_DIR}/${SERVICE}"
        ;;
    *)
        echo "Error: Unknown service '$SERVICE'"
        exit 1
        ;;
esac

# Target directory names (normalized)
TARGET_NAME="${SERVICE}"
case "$SERVICE" in
    orchestrator-service)
        TARGET_NAME="orchestrator"
        ;;
    reasoning-service-go)
        TARGET_NAME="reasoning"
        ;;
    auth-service)
        TARGET_NAME="auth"
        ;;
    gateway-service)
        TARGET_NAME="gateway"
        ;;
    monitor-service)
        TARGET_NAME="monitor"
        ;;
    cluster-service)
        TARGET_NAME="cluster"
        ;;
esac

echo "========================================="
echo "Migrating: $SERVICE -> $TARGET_NAME"
echo "========================================="

# Check if source exists
if [ ! -d "$SOURCE_DIR" ]; then
    echo "Error: Source directory $SOURCE_DIR does not exist"
    exit 1
fi

# Step 1: Move cmd
if [ -d "$SOURCE_DIR/cmd" ]; then
    echo "Step 1: Moving cmd..."
    if [ -d "$SOURCE_DIR/cmd/server" ]; then
        cp -r "$SOURCE_DIR/cmd/server"/* "${ROOT_DIR}/cmd/${TARGET_NAME}/"
    fi
    if [ -d "$SOURCE_DIR/cmd/app" ]; then
        cp -r "$SOURCE_DIR/cmd/app" "${ROOT_DIR}/cmd/${TARGET_NAME}/"
    fi
    echo "  ✓ cmd moved"
else
    echo "  ⚠ No cmd directory found"
fi

# Step 2: Move internal
if [ -d "$SOURCE_DIR/internal" ]; then
    echo "Step 2: Moving internal..."
    cp -r "$SOURCE_DIR/internal"/* "${ROOT_DIR}/internal/${TARGET_NAME}/"
    echo "  ✓ internal moved"
else
    echo "  ⚠ No internal directory found"
fi

# Step 3: Move configs
if [ -d "$SOURCE_DIR/configs" ]; then
    echo "Step 3: Moving configs..."
    mkdir -p "${ROOT_DIR}/configs/${TARGET_NAME}"
    cp -r "$SOURCE_DIR/configs"/* "${ROOT_DIR}/configs/${TARGET_NAME}/"
    echo "  ✓ configs moved"
else
    echo "  ⚠ No configs directory found"
fi

# Step 4: Move Dockerfile
if [ -f "$SOURCE_DIR/Dockerfile" ]; then
    echo "Step 4: Moving Dockerfile..."
    mkdir -p "${ROOT_DIR}/build/docker"
    cp "$SOURCE_DIR/Dockerfile" "${ROOT_DIR}/build/docker/${TARGET_NAME}.Dockerfile"
    echo "  ✓ Dockerfile moved"
else
    echo "  ⚠ No Dockerfile found"
fi

# Step 5: Move pkg (if service has public packages)
if [ -d "$SOURCE_DIR/pkg" ]; then
    echo "Step 5: Moving pkg..."
    # This requires manual review - don't auto-move
    echo "  ⚠ pkg directory found - requires manual review"
    echo "    Consider if these should go to pkg/ (public) or internal/${TARGET_NAME}/ (private)"
fi

# Step 6: Move deployment manifests
if [ -d "$SOURCE_DIR/deployments" ] || [ -d "$SOURCE_DIR/k8s" ]; then
    echo "Step 6: Moving deployment manifests..."
    mkdir -p "${ROOT_DIR}/manifests/base/${TARGET_NAME}"
    if [ -d "$SOURCE_DIR/deployments" ]; then
        cp -r "$SOURCE_DIR/deployments"/* "${ROOT_DIR}/manifests/base/${TARGET_NAME}/"
    fi
    if [ -d "$SOURCE_DIR/k8s" ]; then
        cp -r "$SOURCE_DIR/k8s"/* "${ROOT_DIR}/manifests/base/${TARGET_NAME}/"
    fi
    echo "  ✓ manifests moved"
else
    echo "  ⚠ No deployment manifests found"
fi

echo ""
echo "========================================="
echo "Migration complete!"
echo "========================================="
echo ""
echo "Next steps:"
echo "1. Update import paths in ${ROOT_DIR}/cmd/${TARGET_NAME}/"
echo "2. Update import paths in ${ROOT_DIR}/internal/${TARGET_NAME}/"
echo "3. Test build: go build -o bin/${TARGET_NAME} ./cmd/${TARGET_NAME}"
echo "4. If successful, remove old directory: rm -rf ${SOURCE_DIR}"
echo ""
echo "Import path changes needed:"
echo "  FROM: github.com/kart-io/k8s-agent/${SERVICE}/internal/..."
echo "  TO:   github.com/kart-io/k8s-agent/internal/${TARGET_NAME}/..."
echo ""
