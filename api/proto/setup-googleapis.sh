#!/bin/bash
# setup-googleapis.sh - Setup googleapis dependencies for proto generation

set -e

GOPATH_SRC=$(go env GOPATH)/src
GOOGLEAPIS_DIR="${GOPATH_SRC}/github.com/googleapis/googleapis"
GRPC_GATEWAY_DIR="${GOPATH_SRC}/github.com/grpc-ecosystem/grpc-gateway"

echo "Setting up googleapis dependencies..."

# Clone googleapis if not exists
if [ ! -d "$GOOGLEAPIS_DIR" ]; then
    echo "Cloning googleapis..."
    mkdir -p $(dirname "$GOOGLEAPIS_DIR")
    git clone https://github.com/googleapis/googleapis.git "$GOOGLEAPIS_DIR"
else
    echo "googleapis already exists at $GOOGLEAPIS_DIR"
fi

# Clone grpc-gateway if not exists
if [ ! -d "$GRPC_GATEWAY_DIR" ]; then
    echo "Cloning grpc-gateway..."
    mkdir -p $(dirname "$GRPC_GATEWAY_DIR")
    git clone https://github.com/grpc-ecosystem/grpc-gateway.git "$GRPC_GATEWAY_DIR"
else
    echo "grpc-gateway already exists at $GRPC_GATEWAY_DIR"
fi

echo "Done! You can now run 'make gen-go' to generate proto files."
