# Aetherius Collect Agent

The Aetherius Collect Agent is a lightweight Kubernetes agent that collects events and metrics from clusters and reports them to a central control plane via NATS messaging.

## Features

- **Event Watching**: Monitors Kubernetes events with intelligent filtering for critical issues
- **Metrics Collection**: Collects cluster, node, and pod-level metrics
- **Command Execution**: Safely executes read-only diagnostic commands
- **NATS Communication**: Reliable messaging with automatic reconnection
- **Cloud Detection**: Automatically detects cluster ID from cloud providers (AWS EKS, GCP GKE, Azure AKS)
- **Health Monitoring**: Built-in health checks and Prometheus metrics
- **Security**: Runs as non-root with minimal RBAC permissions

## Architecture

The agent consists of several components:

- **Event Watcher**: Monitors K8s events and filters critical ones
- **Metrics Collector**: Gathers cluster metrics periodically
- **Command Executor**: Executes diagnostic commands safely
- **Communication Manager**: Handles NATS messaging
- **Health Server**: Provides health endpoints

## Configuration

The agent can be configured via YAML file or environment variables:

```yaml
cluster_id: ""                              # Auto-detected if empty
central_endpoint: "nats://central:4222"     # NATS server endpoint
reconnect_delay: 5s
heartbeat_interval: 30s
metrics_interval: 60s
buffer_size: 1000
max_retries: 10
log_level: "info"
enable_metrics: true
enable_events: true
```

### Environment Variables

- `CLUSTER_ID`: Override cluster ID
- `CENTRAL_ENDPOINT`: NATS server endpoint
- `LOG_LEVEL`: Logging level (debug, info, warn, error, fatal)
- `RECONNECT_DELAY`: Delay between reconnection attempts
- `HEARTBEAT_INTERVAL`: Heartbeat send interval
- `METRICS_INTERVAL`: Metrics collection interval
- `ENABLE_METRICS`: Enable metrics collection (true/false)
- `ENABLE_EVENTS`: Enable event watching (true/false)

## Deployment

### Prerequisites

- Kubernetes cluster (v1.20+)
- NATS server accessible from the cluster
- kubectl access to deploy manifests

### Quick Start

1. **Customize Configuration**:
   ```bash
   # Edit the central endpoint in manifests/03-configmap.yaml
   sed -i 's|central.aetherius.io:4222|YOUR-NATS-ENDPOINT:4222|g' manifests/03-configmap.yaml
   ```

2. **Deploy the Agent**:
   ```bash
   kubectl apply -f manifests/
   ```

3. **Verify Deployment**:
   ```bash
   kubectl -n aetherius-agent get pods
   kubectl -n aetherius-agent logs deployment/aetherius-agent
   ```

### Build Custom Image

```bash
# Build the image
docker build -t your-registry/collect-agent:v1.0.0 .

# Push to registry
docker push your-registry/collect-agent:v1.0.0

# Update deployment manifest
sed -i 's|aetherius/collect-agent:v1.0.0|your-registry/collect-agent:v1.0.0|g' manifests/04-deployment.yaml
```

## Development

### Building Locally

```bash
# Install dependencies
go mod download

# Build binary
go build -o collect-agent ./main.go

# Run with config
./collect-agent --config=config.yaml
```

### Testing

```bash
# Run tests
go test ./...

# Test with local cluster
kubectl apply -f manifests/01-namespace.yaml
kubectl apply -f manifests/02-rbac.yaml
kubectl apply -f manifests/03-configmap.yaml

# Run locally (requires kubeconfig)
./collect-agent --config=manifests/03-configmap.yaml
```

## NATS Message Subjects

The agent communicates using the following NATS subjects:

- `agent.register.<cluster_id>` - Agent registration
- `agent.heartbeat.<cluster_id>` - Periodic heartbeat
- `agent.event.<cluster_id>` - Event reports
- `agent.metrics.<cluster_id>` - Metrics reports
- `agent.result.<cluster_id>` - Command execution results
- `agent.command.<cluster_id>` - Commands from central (subscribed)

## Security

- Runs as non-root user (65534:65534)
- Read-only root filesystem
- Minimal RBAC permissions (only read access)
- Command whitelist for safe execution
- No privileged operations

## Monitoring

### Health Endpoints

- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe
- `GET /health/status` - Detailed status JSON
- `GET /metrics` - Prometheus metrics

### Prometheus Metrics

- `agent_running` - Agent running status
- `agent_connected` - NATS connection status
- `agent_uptime_seconds` - Agent uptime
- `agent_*_queue_size` - Queue sizes for different message types

### Logging

Structured JSON logging with configurable levels:

```json
{
  "timestamp": "2025-09-29T10:00:00Z",
  "level": "info",
  "message": "Agent started successfully",
  "cluster_id": "prod-us-west-2"
}
```

## Troubleshooting

### Common Issues

1. **Agent can't connect to NATS**:
   - Check `CENTRAL_ENDPOINT` configuration
   - Verify network connectivity
   - Check NATS server logs

2. **No events being reported**:
   - Check RBAC permissions
   - Verify event filtering logic
   - Check agent logs for errors

3. **High memory usage**:
   - Reduce `buffer_size` in configuration
   - Check for event flooding
   - Monitor queue sizes via `/health/status`

### Debug Mode

Enable debug logging:

```bash
kubectl set env -n aetherius-agent deployment/aetherius-agent LOG_LEVEL=debug
```

### Check Status

```bash
# Get pod status
kubectl -n aetherius-agent get pods

# Check logs
kubectl -n aetherius-agent logs deployment/aetherius-agent --follow

# Check health
kubectl -n aetherius-agent port-forward service/aetherius-agent 8080:8080
curl http://localhost:8080/health/status
```

## License

This project is part of the Aetherius AI Agent system.