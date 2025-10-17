# Version Integration Guide

## Overview

cluster-service now integrates with `github.com/kart-io/version` package to provide comprehensive version management with build-time injection and runtime query capabilities.

## Features

- **Build-time Version Injection**: Version information embedded during compilation via ldflags
- **Multiple Output Formats**: JSON, text, and simplified formats
- **Runtime Query**: Version information available via HTTP API endpoints
- **Build Tracking**: Git commit, branch, tree state, and build date
- **Service Information**: Service name, Go version, compiler, and platform details

## Version Information Structure

```json
{
  "service_name": "cluster-service",
  "git_version": "v1.0.0",
  "git_commit": "abc12345...",
  "git_branch": "master",
  "git_tree_state": "clean",
  "build_date": "2025-10-17T07:14:49Z",
  "go_version": "go1.23.0",
  "compiler": "gc",
  "platform": "linux/amd64"
}
```

## Building with Version Injection

### Using Make (Recommended)

```bash
# Build with automatic version detection
make build

# Display version information that will be injected
make version

# Output:
# Service Name: cluster-service
# Git Version: 2bf9aa4a-dirty
# Git Commit: 2bf9aa4a19771f26e4be3d81b72221c8c68f7b51
# Git Branch: master
# Git Tree State: dirty
# Build Date: 2025-10-17T07:14:49Z
```

### Manual Build

```bash
# Collect version information
SERVICE_NAME="cluster-service"
GIT_VERSION=$(git describe --tags --always --dirty)
GIT_COMMIT=$(git rev-parse HEAD)
GIT_BRANCH=$(git branch --show-current)
GIT_TREE_STATE=$(test -n "`git status --porcelain`" && echo "dirty" || echo "clean")
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

# Build with ldflags
go build -ldflags "
  -X 'github.com/kart-io/version.serviceName=${SERVICE_NAME}'
  -X 'github.com/kart-io/version.gitVersion=${GIT_VERSION}'
  -X 'github.com/kart-io/version.gitCommit=${GIT_COMMIT}'
  -X 'github.com/kart-io/version.gitBranch=${GIT_BRANCH}'
  -X 'github.com/kart-io/version.gitTreeState=${GIT_TREE_STATE}'
  -X 'github.com/kart-io/version.buildDate=${BUILD_DATE}'
" -o bin/cluster-service cmd/server/main.go
```

## Version API Endpoints

### 1. Complete Version Information

**GET** `/version`

Returns full version information in JSON format with response wrapper.

```bash
curl http://localhost:8082/version
```

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "service_name": "cluster-service",
    "git_version": "2bf9aa4a-dirty",
    "git_commit": "2bf9aa4a19771f26e4be3d81b72221c8c68f7b51",
    "git_branch": "master",
    "git_tree_state": "dirty",
    "build_date": "2025-10-17T07:14:49Z",
    "go_version": "go1.23.0",
    "compiler": "gc",
    "platform": "linux/amd64"
  }
}
```

### 2. Simplified Version

**GET** `/version/simple`

Returns only service name and version string.

```bash
curl http://localhost:8082/version/simple
```

**Response:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "service": "cluster-service",
    "version": "2bf9aa4a-dirty"
  }
}
```

### 3. Text Format

**GET** `/version/text`

Returns version information in human-readable table format.

```bash
curl http://localhost:8082/version/text
```

**Response:**
```text
   serviceName: cluster-service
   gitVersion: 2bf9aa4a-dirty
   gitCommit: 2bf9aa4a19771f26e4be3d81b72221c8c68f7b51
   gitBranch: master
   gitTreeState: dirty
   buildDate: 2025-10-17T07:14:49Z
   goVersion: go1.23.0
   compiler: gc
   platform: linux/amd64
```

### 4. Raw JSON Format

**GET** `/version/json`

Returns version information as raw JSON (no response wrapper).

```bash
curl http://localhost:8082/version/json
```

**Response:**
```json
{
  "serviceName": "cluster-service",
  "gitVersion": "2bf9aa4a-dirty",
  "gitCommit": "2bf9aa4a19771f26e4be3d81b72221c8c68f7b51",
  "gitBranch": "master",
  "gitTreeState": "dirty",
  "buildDate": "2025-10-17T07:14:49Z",
  "goVersion": "go1.23.0",
  "compiler": "gc",
  "platform": "linux/amd64"
}
```

## Integration Points

### 1. Application Startup

Version information is automatically logged when the service starts:

```go
// main.go
versionInfo := version.Get()

logger.Infow("Application starting",
    "service", versionInfo.ServiceName,
    "version", versionInfo.GitVersion,
    "commit", versionInfo.GitCommit,
    "branch", versionInfo.GitBranch,
    "build_date", versionInfo.BuildDate,
    "config_path", configPath,
    "enable_k8s_api", enableK8sAPI,
)
```

**Log Output:**
```json
{
  "level": "info",
  "service": "cluster-service",
  "version": "2bf9aa4a-dirty",
  "commit": "2bf9aa4a19771f26e4be3d81b72221c8c68f7b51",
  "branch": "master",
  "build_date": "2025-10-17T07:14:49Z",
  "config_path": "configs/config.yaml",
  "enable_k8s_api": true,
  "message": "Application starting"
}
```

### 2. Logger Initialization

Version information is embedded in logger context:

```go
InitialFields: map[string]interface{}{
    "service": version.Get().ServiceName,
    "version": version.Get().GitVersion,
}
```

This ensures all logs include service and version information for better traceability.

### 3. HTTP Endpoints

Four version endpoints are automatically registered:

- `/version` - Complete version information with response wrapper
- `/version/simple` - Simplified version (service + version only)
- `/version/text` - Human-readable table format
- `/version/json` - Raw JSON format (no wrapper)

## Docker Integration

### Dockerfile Support

Update your Dockerfile to accept build arguments:

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

# Accept build arguments
ARG SERVICE_NAME=cluster-service
ARG VERSION
ARG COMMIT
ARG BRANCH
ARG BUILD_DATE

# Build with version injection
RUN go build -ldflags "\
    -X 'github.com/kart-io/version.serviceName=${SERVICE_NAME}' \
    -X 'github.com/kart-io/version.gitVersion=${VERSION}' \
    -X 'github.com/kart-io/version.gitCommit=${COMMIT}' \
    -X 'github.com/kart-io/version.gitBranch=${BRANCH}' \
    -X 'github.com/kart-io/version.buildDate=${BUILD_DATE}' \
    " -o cluster-service cmd/server/main.go

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/cluster-service /usr/local/bin/

# Set version labels
LABEL version="${VERSION}"
LABEL commit="${COMMIT}"
LABEL branch="${BRANCH}"

ENTRYPOINT ["cluster-service"]
```

### Docker Build with Version

```bash
# Using Make (handles version automatically)
make docker-build

# Manual Docker build
docker build \
  --build-arg SERVICE_NAME=cluster-service \
  --build-arg VERSION=$(git describe --tags --always --dirty) \
  --build-arg COMMIT=$(git rev-parse HEAD) \
  --build-arg BRANCH=$(git branch --show-current) \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t cluster-service:latest .
```

## Testing Version Integration

### 1. Build and Verify

```bash
# Build with version injection
make build

# Check version in logs when starting
./bin/cluster-service -config configs/config.yaml
```

Expected log output:
```json
{
  "level": "info",
  "service": "cluster-service",
  "version": "2bf9aa4a-dirty",
  "commit": "2bf9aa4a19771f26e4be3d81b72221c8c68f7b51",
  "branch": "master",
  "build_date": "2025-10-17T07:14:49Z",
  "message": "Application starting"
}
```

### 2. Test API Endpoints

```bash
# Start the service
./bin/cluster-service -config configs/config.yaml &

# Test complete version endpoint
curl http://localhost:8082/version | jq

# Test simplified version endpoint
curl http://localhost:8082/version/simple | jq

# Test text format
curl http://localhost:8082/version/text

# Test raw JSON format
curl http://localhost:8082/version/json | jq

# Stop the service
kill %1
```

### 3. Automated Testing Script

```bash
#!/bin/bash

BASE_URL="http://localhost:8082"

echo "Testing version endpoints..."

echo -e "\n1. Testing /version (complete):"
curl -s ${BASE_URL}/version | jq -r '.data.git_version'

echo -e "\n2. Testing /version/simple:"
curl -s ${BASE_URL}/version/simple | jq -r '.data.version'

echo -e "\n3. Testing /version/text:"
curl -s ${BASE_URL}/version/text | grep "gitVersion"

echo -e "\n4. Testing /version/json:"
curl -s ${BASE_URL}/version/json | jq -r '.gitVersion'

echo -e "\nAll version endpoints tested successfully!"
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Release

on:
  push:
    tags: ['v*']

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
      with:
        fetch-depth: 0

    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.23

    - name: Build with version info
      run: make build

    - name: Verify version
      run: |
        ./bin/cluster-service -config configs/config.yaml &
        sleep 5
        curl -s http://localhost:8082/version | jq
        kill %1
```

## Best Practices

### 1. Always Use Make for Builds

Using `make build` ensures consistent version injection:

```bash
# ✅ Recommended
make build

# ❌ Not recommended (no version injection)
go build -o bin/cluster-service cmd/server/main.go
```

### 2. Include Version in Monitoring

Use version endpoints for health checks and monitoring:

```bash
# Prometheus exporter can scrape version information
curl http://localhost:8082/version/json
```

### 3. Tag Releases Properly

Follow semantic versioning for Git tags:

```bash
# Create a release tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# Build will automatically use this tag
make build
# Git Version: v1.0.0
```

### 4. Clean vs Dirty Builds

- **Clean**: No uncommitted changes (production builds)
- **Dirty**: Has uncommitted changes (development builds)

```bash
# Check tree state before building
make version
# Git Tree State: clean   ✅ Production ready
# Git Tree State: dirty   ⚠️  Development only
```

## Troubleshooting

### Version Shows as "v0.0.0-dev"

**Cause**: Not in a Git repository or Git not available.

**Solution**:
```bash
# Initialize Git repository
git init
git add .
git commit -m "Initial commit"
```

### Version Shows as "dirty"

**Cause**: Uncommitted changes in working directory.

**Solution**:
```bash
# Commit or stash changes
git add .
git commit -m "Commit message"

# Or stash changes
git stash
```

### Build Date is Incorrect

**Cause**: System time zone issue.

**Solution**: The build date is always in UTC. Ensure your system can execute:
```bash
date -u +'%Y-%m-%dT%H:%M:%SZ'
```

## Reference

- Version Package: `github.com/kart-io/version`
- Version Package Documentation: `../../version/README.md`
- Makefile: `./Makefile`
- Main Entry Point: `./cmd/server/main.go`
- Version Handler: `./internal/handler/version.go`

## Change Log

| Date       | Change                                           |
|------------|--------------------------------------------------|
| 2025-10-17 | Initial version integration with kart-io/version |
| 2025-10-17 | Added 4 version HTTP endpoints                   |
| 2025-10-17 | Updated Makefile with version injection          |
| 2025-10-17 | Integrated version info in logger context        |

---

**Last Updated**: 2025-10-17
**Integration Status**: ✅ Complete
**API Endpoints**: 4 (`/version`, `/version/simple`, `/version/text`, `/version/json`)
