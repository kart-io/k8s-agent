#!/bin/bash

# Test script to send events to orchestrator-service via NATS

NATS_URL="${NATS_URL:-localhost:4222}"
SUBJECT="${1:-internal.event.critical}"

echo "Sending test event to NATS..."
echo "NATS URL: $NATS_URL"
echo "Subject: $SUBJECT"

# Install nats CLI if not available
if ! command -v nats &> /dev/null; then
    echo "Installing NATS CLI..."
    go install github.com/nats-io/natscli/nats@latest
fi

# Test event payload
cat <<EOF | nats pub "$SUBJECT" --server="nats://$NATS_URL"
{
  "type": "pod_failure",
  "cluster_id": "test-cluster-001",
  "severity": "critical",
  "payload": {
    "namespace": "default",
    "pod_name": "nginx-pod-12345",
    "container": "nginx",
    "reason": "CrashLoopBackOff",
    "message": "Container exited with code 1",
    "restart_count": 5
  },
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}
EOF

echo ""
echo "Event sent successfully!"
echo ""
echo "To monitor events, run:"
echo "  nats sub 'internal.event.*' --server='nats://$NATS_URL'"
