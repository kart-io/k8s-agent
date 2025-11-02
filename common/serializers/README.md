# Serializers Package

The `serializers` package provides a unified serialization interface and multiple format implementations for data serialization/deserialization.

## Overview

This package defines a common `Serializer` interface and provides implementations for popular serialization formats:

- **JSON**: Human-readable, widely supported
- **MessagePack**: Binary format, fast and compact
- **YAML**: Human-readable, ideal for configuration files

## Interface

```go
type Serializer interface {
    // Marshal serializes a value to bytes.
    Marshal(v interface{}) ([]byte, error)

    // Unmarshal deserializes bytes to a value.
    Unmarshal(data []byte, v interface{}) error

    // Name returns the serializer name for logging and metrics.
    Name() string
}
```

## Available Serializers

### JSON Serializer

**Location**: `common/serializers/json.go`

Uses Go's standard `encoding/json` package. Best for:
- API responses
- Human-readable data
- Wide language compatibility

**Example**:
```go
import "github.com/kart-io/k8s-agent/common/serializers"

serializer := serializers.NewJSONSerializer()
data, err := serializer.Marshal(myStruct)
```

**Performance** (Apple M4 Pro):
- Marshal: ~169 ns/op, 176 B/op
- Unmarshal: ~772 ns/op, 296 B/op

### MessagePack Serializer

**Location**: `common/serializers/msgpack.go`

Uses `github.com/vmihailenco/msgpack/v5`. Best for:
- High-performance applications
- Network protocols
- Storage optimization (19.63% smaller than JSON)

**Example**:
```go
import "github.com/kart-io/k8s-agent/common/serializers"

serializer := serializers.NewMsgpackSerializer()
data, err := serializer.Marshal(myStruct)
```

**Performance** (Apple M4 Pro):
- Marshal: ~190 ns/op, 304 B/op
- Unmarshal: ~279 ns/op, 108 B/op
- **2.76x faster unmarshal than JSON**

### YAML Serializer

**Location**: `common/serializers/yaml.go`

Uses `gopkg.in/yaml.v3`. Best for:
- Configuration files
- Human-readable data with complex structures
- CI/CD manifests

**Example**:
```go
import "github.com/kart-io/k8s-agent/common/serializers"

serializer := serializers.NewYAMLSerializer()
data, err := serializer.Marshal(myStruct)
```

**Performance** (Apple M4 Pro):
- Marshal: ~3582 ns/op, 16768 B/op
- Unmarshal: ~5400 ns/op, 10000 B/op
- **Trade-off**: Slower but human-readable

## Performance Comparison

| Serializer | Marshal (ns/op) | Unmarshal (ns/op) | Size vs JSON | Use Case |
|------------|-----------------|-------------------|--------------|----------|
| JSON       | 169             | 772               | Baseline     | API, compatibility |
| MessagePack| 190             | 279               | -19.63%      | Performance, storage |
| YAML       | 3,582           | 5,400             | ~varies      | Configuration, readability |

**Recommendation**:
- **Default**: JSON (good balance of speed and compatibility)
- **Performance**: MessagePack (2.76x faster unmarshal, 19.63% smaller)
- **Configuration**: YAML (human-readable, ideal for config files)

## Usage in Cache Package

The cache package (especially L2Cache) uses serializers for encoding/decoding values:

```go
import (
    "github.com/kart-io/k8s-agent/common/cache"
    "github.com/kart-io/k8s-agent/common/cache/l2"
    "github.com/kart-io/k8s-agent/common/serializers"
)

// Create L2 cache with MessagePack serializer
l2Cache, err := l2.NewL2Cache[MyType](remoteCache,
    cache.WithSerializer(serializers.NewMsgpackSerializer()),
    cache.WithLocalSize(10000),
)
```

## Testing

Run tests:
```bash
cd common
go test ./serializers/... -v
```

Run benchmarks:
```bash
cd common
go test ./serializers/... -bench=. -benchmem
```

## Migration from `common/cache/serializers`

**Old import** (deprecated):
```go
import "github.com/kart-io/k8s-agent/common/cache/serializers"
```

**New import** (recommended):
```go
import "github.com/kart-io/k8s-agent/common/serializers"
```

The cache package provides a type alias for backward compatibility, but new code should import directly from `common/serializers`.
