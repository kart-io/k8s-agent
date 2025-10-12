# Auth Service - Developer Quickstart Guide

**Last Updated**: 2025-10-09
**Target Audience**: Backend developers implementing or extending the auth-service

## Prerequisites

- Go 1.21 or later
- PostgreSQL 12+ (running locally or via Docker)
- Redis 6+ (running locally or via Docker)
- Git
- Make (optional, but recommended)

## Quick Start (5 minutes)

### 1. Clone and Navigate

```bash
cd /path/to/k8s-agent
cd auth-service
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Setup Databases

**Option A: Using Docker Compose** (Recommended)

```bash
# Create docker-compose.yml in auth-service directory
cat > docker-compose.yml <<EOF
version: '3.8'
services:
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_DB: k8s_agent_auth
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
EOF

# Start databases
docker-compose up -d
```

**Option B: Local Installation**

```bash
# macOS
brew install postgresql redis
brew services start postgresql
brew services start redis

# Ubuntu/Debian
sudo apt install postgresql redis-server
sudo systemctl start postgresql redis-server

# Create database
createdb k8s_agent_auth
```

### 4. Configure Application

```bash
# Copy example config
cp configs/config.yaml.example configs/config.yaml

# Edit configuration (or use defaults)
vim configs/config.yaml
```

**Default Configuration** (`configs/config.yaml`):

```yaml
server:
  port: 8090
  mode: debug  # debug, release, test

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: k8s_agent_auth
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 5

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 10

jwt:
  secret: "change-this-in-production"
  expires_hours: 24

logging:
  level: info  # debug, info, warn, error
  format: json  # json, text
```

### 5. Run Application

```bash
# Using Go
go run cmd/server/main.go

# Or using Make
make run

# Or build and run
make build
./bin/auth-service
```

You should see:

```
INFO Server started on port 8090
INFO Database connected successfully
INFO Redis connected successfully
INFO Migrations completed
INFO Default admin user created: admin/admin123
```

### 6. Test the API

```bash
# Health check
curl http://localhost:8090/health

# Login as admin
curl -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Save the token from response
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Get current user info
curl http://localhost:8090/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"

# Get user menus
curl http://localhost:8090/api/v1/auth/menus \
  -H "Authorization: Bearer $TOKEN"
```

🎉 **Congratulations!** The auth-service is now running.

## Project Structure Explained

```
auth-service/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
│
├── internal/                     # Private application code
│   ├── handler/                  # HTTP handlers (controllers)
│   │   ├── auth_handler.go       # Login, logout, user info
│   │   ├── user_handler.go       # User CRUD
│   │   ├── role_handler.go       # Role CRUD
│   │   └── permission_handler.go # Permission CRUD
│   │
│   ├── middleware/               # HTTP middleware
│   │   ├── jwt.go                # JWT authentication
│   │   ├── permission.go         # Permission checks
│   │   ├── api_key.go            # API key auth
│   │   └── cors.go               # CORS handling
│   │
│   ├── model/                    # Database models (GORM)
│   │   ├── user.go
│   │   ├── role.go
│   │   ├── permission.go
│   │   └── api_key.go
│   │
│   ├── service/                  # Business logic
│   │   ├── auth_service.go       # Authentication logic
│   │   ├── user_service.go       # User management
│   │   ├── role_service.go       # Role management
│   │   └── permission_service.go # Permission management
│   │
│   └── storage/                  # Data access layer
│       ├── postgres.go           # DB connection
│       ├── redis.go              # Redis connection
│       └── migrate.go            # Migrations
│
├── pkg/
│   └── types/                    # Shared types
│       └── types.go
│
├── configs/
│   └── config.yaml               # Configuration file
│
├── tests/
│   ├── integration/              # Integration tests
│   └── unit/                     # Unit tests
│
├── scripts/                      # Utility scripts
├── Dockerfile
├── Makefile
└── go.mod
```

## Development Workflow

### Adding a New Endpoint

**Example**: Add "Change Password" feature

**Step 1: Define the request/response types** (`pkg/types/types.go`)

```go
type ChangePasswordRequest struct {
    OldPassword string `json:"old_password" binding:"required"`
    NewPassword string `json:"new_password" binding:"required,min=8"`
}
```

**Step 2: Add service method** (`internal/service/auth_service.go`)

```go
func (s *AuthService) ChangePassword(userID string, req *types.ChangePasswordRequest) error {
    // 1. Get user from database
    // 2. Verify old password
    // 3. Hash new password
    // 4. Update database
    // 5. Invalidate existing tokens (add to Redis blacklist)
}
```

**Step 3: Add handler** (`internal/handler/auth_handler.go`)

```go
func (h *AuthHandler) ChangePassword(c *gin.Context) {
    var req types.ChangePasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    userID := c.GetString("user_id") // From JWT middleware
    if err := h.authService.ChangePassword(userID, &req); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"message": "Password changed successfully"})
}
```

**Step 4: Register route** (`cmd/server/main.go`)

```go
authenticated.POST("/auth/change-password", authHandler.ChangePassword)
```

**Step 5: Write tests** (`tests/integration/auth_test.go`)

```go
func TestChangePassword(t *testing.T) {
    // Test successful password change
    // Test with wrong old password
    // Test with weak new password
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -v ./internal/service -run TestAuthService_Login

# Run integration tests only
go test ./tests/integration/...
```

### Database Migrations

**Auto Migration** (Development):

The application runs auto-migration on startup using GORM.

**Manual Migration** (Production):

```bash
# TODO: Implement proper migration tool (e.g., golang-migrate)
# For now, migrations are in internal/storage/migrate.go
```

**Add New Table/Column**:

1. Update model in `internal/model/`
2. Add migration logic in `internal/storage/migrate.go`
3. Test migration on development database
4. Document schema changes in `/specs/001-auth-service-a/data-model.md`

### Debugging

**Enable Debug Logging**:

```yaml
# configs/config.yaml
server:
  mode: debug

logging:
  level: debug
  format: text  # More readable in development
```

**Using Delve Debugger**:

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Run with debugger
dlv debug cmd/server/main.go

# Set breakpoint
(dlv) break internal/handler/auth_handler.go:45
(dlv) continue
```

**Logging Best Practices**:

```go
// Use structured logging
logger.WithFields(logrus.Fields{
    "user_id": userID,
    "action": "login",
}).Info("User logged in successfully")

// Log errors with context
logger.WithError(err).WithFields(logrus.Fields{
    "user_id": userID,
}).Error("Failed to change password")
```

## Common Development Tasks

### 1. Add New Permission

```sql
-- Connect to database
psql k8s_agent_auth

-- Insert new permission
INSERT INTO permissions (id, parent_id, name, code, type, path, method, sort, status)
VALUES (
  gen_random_uuid(),
  'parent-uuid',
  'Export Users',
  'user:export',
  'api',
  '/api/v1/users/export',
  'GET',
  10,
  1
);
```

### 2. Create New Role

```go
// Via API
curl -X POST http://localhost:8090/api/v1/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Content Manager",
    "code": "content_manager",
    "description": "Manages content"
  }'
```

### 3. Assign Permissions to Role

```go
// Via API
curl -X POST http://localhost:8090/api/v1/roles/{role-id}/permissions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "permission_ids": ["uuid1", "uuid2", "uuid3"]
  }'
```

### 4. Generate API Key

```go
// Via API
curl -X POST http://localhost:8090/api/v1/api-keys \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Service Integration Key",
    "description": "For microservice communication",
    "expires_at": "2025-12-31T23:59:59Z"
  }'

// Response includes key and secret (shown only once)
{
  "key": "ak_1234567890abcdef",
  "secret": "sk_abcdefghijklmnop",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

## Integration with Frontend

### Vue.js Example

```javascript
// auth.js - Authentication service
import axios from 'axios';

const API_BASE = 'http://localhost:8090/api/v1';

export async function login(username, password) {
  const response = await axios.post(`${API_BASE}/auth/login`, {
    username,
    password
  });

  // Save token
  localStorage.setItem('token', response.data.token);
  localStorage.setItem('user', JSON.stringify(response.data.user));

  return response.data;
}

export async function getMenus() {
  const token = localStorage.getItem('token');
  const response = await axios.get(`${API_BASE}/auth/menus`, {
    headers: { Authorization: `Bearer ${token}` }
  });
  return response.data;
}

export function hasPermission(permissionCode) {
  const user = JSON.parse(localStorage.getItem('user'));
  // Check user.permissions array
  return user.permissions.some(p => p.code === permissionCode);
}
```

### Request Interceptor

```javascript
// axios-config.js
import axios from 'axios';
import router from './router';

axios.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

axios.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      // Token expired, redirect to login
      localStorage.clear();
      router.push('/login');
    }
    return Promise.reject(error);
  }
);
```

## Production Deployment

### Building for Production

```bash
# Build binary
make build

# Build Docker image
make docker-build

# Or manually
docker build -t auth-service:v1.0.0 .
```

### Configuration for Production

```yaml
# configs/config.yaml (production)
server:
  port: 8090
  mode: release

database:
  host: postgres.prod.internal
  port: 5432
  user: auth_service
  password: ${DB_PASSWORD}  # From environment variable
  dbname: k8s_agent_auth_prod
  sslmode: require
  max_open_conns: 100
  max_idle_conns: 25

redis:
  host: redis.prod.internal
  port: 6379
  password: ${REDIS_PASSWORD}
  db: 0
  pool_size: 50

jwt:
  secret: ${JWT_SECRET}  # MUST be from environment variable
  expires_hours: 24

logging:
  level: info
  format: json  # For log aggregation
```

### Environment Variables

```bash
export DB_PASSWORD="strong-db-password"
export REDIS_PASSWORD="strong-redis-password"
export JWT_SECRET="very-strong-jwt-secret-at-least-32-chars"
```

### Docker Deployment

```dockerfile
# Dockerfile is already provided
docker run -d \
  --name auth-service \
  -p 8090:8090 \
  -e DB_PASSWORD=$DB_PASSWORD \
  -e REDIS_PASSWORD=$REDIS_PASSWORD \
  -e JWT_SECRET=$JWT_SECRET \
  -v /path/to/config.yaml:/app/configs/config.yaml \
  auth-service:v1.0.0
```

## Troubleshooting

### Database Connection Fails

```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Check database exists
psql -l | grep k8s_agent_auth

# Check connection manually
psql -h localhost -p 5432 -U postgres -d k8s_agent_auth
```

### Redis Connection Fails

```bash
# Check Redis is running
redis-cli ping  # Should return PONG

# Check Redis connection
redis-cli -h localhost -p 6379
```

### JWT Token Issues

```bash
# Decode JWT token (without verification)
echo $TOKEN | cut -d'.' -f2 | base64 -d | jq

# Check token expiration
# Look for "exp" field in decoded token
```

### Permission Denied Errors

```bash
# Check user roles
SELECT u.username, r.name, r.code
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
WHERE u.username = 'admin';

# Check role permissions
SELECT r.name, p.code, p.type
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE r.code = 'admin';
```

## Next Steps

1. ✅ Complete remaining API endpoints (if any)
2. ✅ Write comprehensive integration tests
3. ✅ Add API documentation (Swagger UI)
4. ✅ Implement rate limiting
5. ✅ Add audit logging
6. ✅ Implement token refresh mechanism
7. ✅ Add OAuth 2.0 support (future)
8. ✅ Add MFA support (future)

## Additional Resources

- [Project README](../../auth-service/README.md)
- [Data Model Documentation](./data-model.md)
- [API Contracts](./contracts/)
- [Research & Technology Decisions](./research.md)
- [Implementation Plan](./plan.md)

## Getting Help

- Check existing issues in the repository
- Review logs for error details
- Consult the data model documentation
- Reach out to the development team

---

**Happy Coding!** 🚀
