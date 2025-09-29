#!/bin/bash

# Aetherius Collect Agent Deployment Script
# Usage: ./deploy.sh [cluster-id] [central-endpoint]

set -e

CLUSTER_ID=${1:-""}
CENTRAL_ENDPOINT=${2:-"nats://central.aetherius.io:4222"}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFESTS_DIR="$SCRIPT_DIR/../manifests"

echo "=== Deploying Aetherius Collect Agent ==="
echo "Cluster ID: ${CLUSTER_ID:-"auto-detect"}"
echo "Central Endpoint: $CENTRAL_ENDPOINT"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed or not in PATH"
    exit 1
fi

# Check if we can access the cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "Error: Cannot access Kubernetes cluster. Please check your kubeconfig."
    exit 1
fi

echo ""
echo "Current cluster context:"
kubectl config current-context

read -p "Continue with deployment? (y/N) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Deployment cancelled."
    exit 0
fi

# Deploy namespace and RBAC first
echo ""
echo "=== Creating namespace and RBAC ==="
kubectl apply -f "$MANIFESTS_DIR/01-namespace.yaml"
kubectl apply -f "$MANIFESTS_DIR/02-rbac.yaml"

# Create or update ConfigMap with custom values
echo ""
echo "=== Creating ConfigMap ==="
TEMP_CONFIGMAP="/tmp/agent-configmap.yaml"
cp "$MANIFESTS_DIR/03-configmap.yaml" "$TEMP_CONFIGMAP"

# Update cluster_id if provided
if [ -n "$CLUSTER_ID" ]; then
    sed -i.bak "s|cluster_id: \"\"|cluster_id: \"$CLUSTER_ID\"|g" "$TEMP_CONFIGMAP"
fi

# Update central_endpoint
sed -i.bak "s|central_endpoint: \".*\"|central_endpoint: \"$CENTRAL_ENDPOINT\"|g" "$TEMP_CONFIGMAP"

kubectl apply -f "$TEMP_CONFIGMAP"
rm -f "$TEMP_CONFIGMAP" "$TEMP_CONFIGMAP.bak" 2>/dev/null || true

# Deploy the agent
echo ""
echo "=== Deploying agent ==="
kubectl apply -f "$MANIFESTS_DIR/04-deployment.yaml"
kubectl apply -f "$MANIFESTS_DIR/05-service.yaml"

# Wait for deployment
echo ""
echo "=== Waiting for deployment to be ready ==="
kubectl -n aetherius-agent rollout status deployment/aetherius-agent --timeout=300s

# Show status
echo ""
echo "=== Deployment Status ==="
kubectl -n aetherius-agent get pods -l app.kubernetes.io/name=aetherius-agent
kubectl -n aetherius-agent get svc aetherius-agent

# Show recent logs
echo ""
echo "=== Recent Logs ==="
kubectl -n aetherius-agent logs deployment/aetherius-agent --tail=20

# Health check
echo ""
echo "=== Health Check ==="
AGENT_POD=$(kubectl -n aetherius-agent get pods -l app.kubernetes.io/name=aetherius-agent -o jsonpath='{.items[0].metadata.name}')
if [ -n "$AGENT_POD" ]; then
    echo "Agent Pod: $AGENT_POD"
    kubectl -n aetherius-agent exec "$AGENT_POD" -- wget -q -O- http://localhost:8080/health/status | jq '.' 2>/dev/null || kubectl -n aetherius-agent exec "$AGENT_POD" -- wget -q -O- http://localhost:8080/health/status
else
    echo "Warning: Could not find agent pod for health check"
fi

echo ""
echo "=== Deployment Complete ==="
echo ""
echo "Useful commands:"
echo "  View logs:    kubectl -n aetherius-agent logs deployment/aetherius-agent --follow"
echo "  Check status: kubectl -n aetherius-agent get pods"
echo "  Port forward: kubectl -n aetherius-agent port-forward service/aetherius-agent 8080:8080"
echo "  Delete:       kubectl delete namespace aetherius-agent"
echo ""
echo "Health endpoints (after port-forward):"
echo "  http://localhost:8080/health/live"
echo "  http://localhost:8080/health/ready"
echo "  http://localhost:8080/health/status"
echo "  http://localhost:8080/metrics"