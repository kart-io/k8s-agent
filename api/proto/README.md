# Protos

This directory contains all Protocol Buffer (protobuf) definitions for the k8s-agent project.

## Directory Structure

```
protos/
├── agentmanager/           # Agent Manager service protos
│   ├── agent/v1/          # Agent management APIs
│   ├── command/v1/        # Command execution APIs
│   └── event/v1/          # Event handling APIs
├── common/                # Common/shared protos
│   ├── example/v1/        # Example service
│   └── health/v1/         # Health check service
└── gen/                   # Generated Go code (gitignored)
```

## Prerequisites

Before generating proto files, you need to install:

1. **Protocol Buffers Compiler (protoc)**
   ```bash
   # macOS
   brew install protobuf

   # Linux
   apt-get install -y protobuf-compiler

   # Or download from: https://github.com/protocolbuffers/protobuf/releases
   ```

2. **Go Plugins**
   ```bash
   make install-tools
   ```

   This will install:
   - `protoc-gen-go` - Go code generator
   - `protoc-gen-go-grpc` - Go gRPC code generator
   - `protoc-gen-grpc-gateway` - gRPC-Gateway generator
   - `protoc-gen-openapiv2` - OpenAPI/Swagger generator

## Usage

### Generate All Proto Files

```bash
cd protos
make gen-go
```

This will generate Go code for all proto files in the `gen/` directory.

### Generate Specific Services

```bash
# Generate only agent-manager protos
make gen-agentmanager

# Generate only common protos
make gen-common
```

### Validate Proto Files

```bash
make validate
```

### List All Proto Files

```bash
make list
```

### Clean Generated Files

```bash
make clean
```

### Show Help

```bash
make help
```

## Generated Code

Generated Go code is placed in the `gen/` directory with the following structure:

```
gen/
├── agentmanager/
│   ├── agent/v1/
│   │   ├── agent.pb.go           # Message definitions
│   │   ├── agent_grpc.pb.go      # gRPC service definitions
│   │   └── agent.pb.gw.go        # gRPC-Gateway reverse proxy
│   ├── command/v1/
│   └── event/v1/
└── common/
    ├── example/v1/
    └── health/v1/
```

## Import Path Convention

All proto files use the following Go package convention:

```protobuf
option go_package = "github.com/kart-io/k8s-agent/protos/gen/<service>/<api>/<version>;<package>";
```

Example:
```protobuf
option go_package = "github.com/kart-io/k8s-agent/protos/gen/agentmanager/agent/v1;agentv1";
```

## Adding New Proto Files

1. Create the proto file in the appropriate directory:
   ```
   protos/<service>/<api>/v1/your_service.proto
   ```

2. Define the Go package in your proto file:
   ```protobuf
   syntax = "proto3";

   package <service>.<api>.v1;

   option go_package = "github.com/kart-io/k8s-agent/protos/gen/<service>/<api>/v1;<package_name>";
   ```

3. Update the Makefile to include your new proto file:
   - Add directory creation in the appropriate `gen-*` target
   - Add protoc generation command

4. Generate the code:
   ```bash
   make gen-go
   ```

## gRPC-Gateway

The agent-manager services use gRPC-Gateway to expose gRPC services as REST APIs. The gateway mappings are defined in the proto files using annotations:

```protobuf
import "google/api/annotations.proto";

service AgentService {
  rpc ListAgents(ListAgentsRequest) returns (ListAgentsResponse) {
    option (google.api.http) = {
      get: "/api/v1/grpc/agents"
    };
  }
}
```

## Version Policy

- All APIs follow semantic versioning (v1, v2, etc.)
- Breaking changes require a new version
- Backward-compatible changes can be added to existing versions
- Deprecated fields should be marked with `[deprecated = true]`

## Common Patterns

### Message Naming
- Request messages: `<Operation>Request`
- Response messages: `<Operation>Response`
- Resources: Use singular nouns (e.g., `Agent`, not `Agents`)

### Field Naming
- Use `snake_case` for field names
- Use descriptive names (e.g., `cluster_id` instead of `cid`)

### Service Naming
- Use `<Resource>Service` pattern
- Group related operations in the same service

## Troubleshooting

### "protoc-gen-go: program not found or is not executable"

Make sure `$(go env GOPATH)/bin` is in your PATH:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### "google/api/annotations.proto: File not found"

Install googleapis protos:
```bash
git clone https://github.com/googleapis/googleapis.git $(go env GOPATH)/src/github.com/googleapis/googleapis
```

Or use buf.build for dependency management.

## References

- [Protocol Buffers Documentation](https://protobuf.dev/)
- [gRPC-Go Documentation](https://grpc.io/docs/languages/go/)
- [gRPC-Gateway Documentation](https://grpc-ecosystem.github.io/grpc-gateway/)
- [Google API Design Guide](https://cloud.google.com/apis/design)
