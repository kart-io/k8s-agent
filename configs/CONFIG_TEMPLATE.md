# Configuration Template for Aetherius Services
# This file provides a template for creating service configurations
# Based on OneX v2 patterns

## Service Configuration Template

Each service should have the following configuration files:

```
configs/<service-name>/
├── config.yaml              # Default/base configuration
├── config-dev.yaml          # Development environment
├── config-test.yaml         # Testing environment
├── config-staging.yaml      # Staging environment
├── config-prod.yaml         # Production environment
└── config-local.yaml        # Local development override
```

## Configuration File Structure

### Basic Service Configuration

```yaml
# Service Information
service:
  name: "service-name"
  version: "v1.0.0"
  environment: "development"  # development, testing, staging, production

# Server Configuration
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"  # debug, release
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 10s

# Logging Configuration
log:
  level: "info"  # debug, info, warn, error
  format: "json"  # json, text
  output: "stdout"  # stdout, stderr, file
  file:
    path: "/var/log/aetherius/service.log"
    max_size: 100  # MB
    max_backups: 10
    max_age: 30  # days
    compress: true

# Database Configuration (MySQL)
database:
  host: "localhost"
  port: 3306
  username: "aetherius"
  password: "aetherius_pwd"
  database: "aetherius"
  charset: "utf8mb4"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600s
  log_mode: false

# Redis Configuration
redis:
  host: "localhost"
  port: 6379
  password: ""
  database: 0
  max_retries: 3
  pool_size: 10
  min_idle_conns: 5

# NATS Configuration
nats:
  url: "nats://localhost:4222"
  cluster_id: "aetherius-cluster"
  client_id: "service-client"
  max_reconnects: 10
  reconnect_wait: 2s

# gRPC Configuration
grpc:
  enabled: true
  port: 9090
  max_recv_msg_size: 4194304  # 4MB
  max_send_msg_size: 4194304  # 4MB
  keepalive:
    time: 30s
    timeout: 10s

# Metrics Configuration
metrics:
  enabled: true
  port: 9091
  path: "/metrics"

# Health Check Configuration
health:
  enabled: true
  port: 9092
  path: "/health"
  liveness_path: "/health/live"
  readiness_path: "/health/ready"

# Tracing Configuration (OpenTelemetry)
tracing:
  enabled: false
  endpoint: "http://localhost:4317"
  service_name: "service-name"
  sample_rate: 0.1

# Feature Flags
features:
  enable_debug: false
  enable_profiling: false
  enable_swagger: false
```

### Environment-Specific Overrides

#### Development (config-dev.yaml)
```yaml
server:
  mode: "debug"

log:
  level: "debug"
  format: "text"

database:
  log_mode: true

features:
  enable_debug: true
  enable_profiling: true
  enable_swagger: true
```

#### Testing (config-test.yaml)
```yaml
server:
  mode: "release"

log:
  level: "info"

database:
  database: "aetherius_test"
```

#### Staging (config-staging.yaml)
```yaml
server:
  mode: "release"

log:
  level: "info"
  format: "json"

database:
  max_open_conns: 50

features:
  enable_debug: false
  enable_profiling: false
```

#### Production (config-prod.yaml)
```yaml
server:
  mode: "release"

log:
  level: "warn"
  format: "json"
  output: "file"

database:
  max_open_conns: 200
  max_idle_conns: 20

redis:
  pool_size: 50
  min_idle_conns: 10

tracing:
  enabled: true
  sample_rate: 0.01

features:
  enable_debug: false
  enable_profiling: false
  enable_swagger: false
```

## Configuration Loading Priority

1. **Environment Variables** (highest priority)
2. **Command-line Flags**
3. **Environment-specific config** (config-{env}.yaml)
4. **Base config** (config.yaml)
5. **Default values** (in code)

## Environment Variable Mapping

Environment variables can override configuration values:

```bash
# Service
AETHERIUS_SERVICE_NAME=my-service
AETHERIUS_SERVICE_ENVIRONMENT=production

# Server
AETHERIUS_SERVER_HOST=0.0.0.0
AETHERIUS_SERVER_PORT=8080

# Database
AETHERIUS_DB_HOST=db.example.com
AETHERIUS_DB_PORT=3306
AETHERIUS_DB_USERNAME=user
AETHERIUS_DB_PASSWORD=secret
AETHERIUS_DB_DATABASE=dbname

# Redis
AETHERIUS_REDIS_HOST=redis.example.com
AETHERIUS_REDIS_PORT=6379
AETHERIUS_REDIS_PASSWORD=secret

# NATS
AETHERIUS_NATS_URL=nats://nats.example.com:4222
```

## Security Best Practices

1. **Never commit secrets** to version control
2. **Use environment variables** for sensitive data in production
3. **Use secret management systems** (Vault, K8s Secrets, AWS Secrets Manager)
4. **Encrypt sensitive config files** at rest
5. **Rotate credentials** regularly
6. **Use minimal permissions** for database users

## Configuration Validation

Each service should validate its configuration on startup:

```go
func ValidateConfig(cfg *Config) error {
    if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
        return errors.New("invalid server port")
    }

    if cfg.Database.MaxOpenConns < 1 {
        return errors.New("database max_open_conns must be positive")
    }

    if cfg.Log.Level == "" {
        return errors.New("log level is required")
    }

    return nil
}
```

## Example Usage

### Loading Configuration in Go

```go
import (
    "github.com/spf13/viper"
)

func LoadConfig(configPath string, env string) (*Config, error) {
    v := viper.New()

    // Set config file
    v.SetConfigFile(configPath)

    // Read base config
    if err := v.ReadInConfig(); err != nil {
        return nil, err
    }

    // Merge environment-specific config
    envConfig := fmt.Sprintf("config-%s.yaml", env)
    v.SetConfigFile(envConfig)
    v.MergeInConfig()

    // Enable environment variable override
    v.SetEnvPrefix("AETHERIUS")
    v.AutomaticEnv()

    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    // Validate configuration
    if err := ValidateConfig(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

### Using Configuration

```go
func main() {
    // Load config
    cfg, err := LoadConfig("configs/agent-manager/config.yaml", "production")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Use config
    server := &http.Server{
        Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
    }

    // Start server
    log.Fatal(server.ListenAndServe())
}
```

## Configuration Testing

Always test configuration loading:

```go
func TestLoadConfig(t *testing.T) {
    cfg, err := LoadConfig("testdata/config.yaml", "test")
    assert.NoError(t, err)
    assert.Equal(t, "test", cfg.Service.Environment)
    assert.Equal(t, 8080, cfg.Server.Port)
}
```

## Documentation

Each service's configuration should be documented in:
- `configs/<service>/README.md` - Configuration guide
- `docs/configuration.md` - Global configuration documentation
- API documentation - Configuration endpoints (if applicable)

## Migration Guide

When changing configuration schema:

1. **Update this template**
2. **Update all environment configs**
3. **Add migration notes** to CHANGELOG.md
4. **Update documentation**
5. **Provide backward compatibility** when possible
6. **Version the configuration schema**

## Related Files

- Configuration loading code: `internal/config/`
- Environment variable mapping: `.env.example`
- Default values: `internal/config/defaults.go`
- Validation: `internal/config/validation.go`
