#!/bin/bash
# Aetherius Collect Agent Deployment Script
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
NAMESPACE="aetherius-agent"
CLUSTER_ID="${CLUSTER_ID:-}"
CENTRAL_ENDPOINT="${CENTRAL_ENDPOINT:-nats://central.aetherius.io:4222}"
IMAGE_TAG="${IMAGE_TAG:-v1.0.0}"

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

log_info "Deploying Aetherius Collect Agent..."
log_info "Cluster ID: ${CLUSTER_ID:-auto-detect}"
log_info "Central Endpoint: $CENTRAL_ENDPOINT"
log_info "Image Tag: $IMAGE_TAG"

# Apply manifests
kubectl apply -f manifests/

log_info "Waiting for deployment..."
kubectl -n $NAMESPACE rollout status deployment/aetherius-agent --timeout=300s

log_info "Deployment complete ✓"
kubectl -n $NAMESPACE get pods
