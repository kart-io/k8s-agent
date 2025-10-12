# Configuration Files

This directory contains configuration files for the cluster-service.

## Files

- **config.yaml** - Default configuration
- **config.dev.yaml** - Development environment configuration
- **config.prod.yaml** - Production environment configuration

## Usage

### Default Configuration

```bash
./server
# or
./server -config configs/config.yaml
```

### Development Environment

```bash
./server -config configs/config.dev.yaml
```

### Production Environment

```bash
./server -config configs/config.prod.yaml
```

## Configuration Structure

```yaml
server:
  port: 8082              # HTTP server port
  mode: debug             # Gin mode: debug, release, test
  read_timeout: 10s       # Request read timeout
  write_timeout: 10s      # Response write timeout

database:
  host: localhost         # PostgreSQL host
  port: 5432             # PostgreSQL port
  user: postgres         # Database user
  password: postgres     # Database password
  dbname: cluster_db     # Database name
  sslmode: disable       # SSL mode: disable, require, verify-ca, verify-full
  max_open_conns: 25     # Maximum open connections
  max_idle_conns: 5      # Maximum idle connections

jwt:
  secret: your-secret    # JWT signing secret

logging:
  level: info           # Log level: debug, info, warn, error
  format: json          # Log format: json, text
```

## Environment Variables

For production, you can use environment variables:

```bash
export DB_USER=myuser
export DB_PASSWORD=mypassword
export JWT_SECRET=my-super-secret-key
```

Then reference them in config:

```yaml
database:
  user: ${DB_USER}
  password: ${DB_PASSWORD}

jwt:
  secret: ${JWT_SECRET}
```

## Security Notes

1. **Never commit sensitive data** to version control
2. Use environment variables for secrets in production
3. Keep JWT secret long and random (minimum 32 characters)
4. Use SSL/TLS for database connections in production
5. Consider using secrets management tools (Vault, AWS Secrets Manager, etc.)

## Local Overrides

Create `config.local.yaml` for local development overrides (gitignored):

```yaml
database:
  password: my-local-password

logging:
  level: debug
```
