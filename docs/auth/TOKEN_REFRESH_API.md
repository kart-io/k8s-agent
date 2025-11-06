# Token Refresh API Documentation

## Overview

The Token Refresh API implements JWT token refresh mechanism with enhanced security features including:

- Refresh token rotation (old token is revoked when new one is issued)
- Refresh token blacklisting to prevent replay attacks
- Redis-based refresh token storage
- 7-day refresh token validity (vs 2-hour access token)
- User validation on every refresh

## Endpoints

### POST /api/v1/auth/refresh

Refreshes an expired or soon-to-expire access token using a valid refresh token.

**Authentication Required**: No (uses refresh token in request body)

#### Request

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Response (200 OK)

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-11-06T10:30:00Z",
    "expires_in": 7200
  }
}
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| access_token | string | New JWT access token (2 hour validity) |
| refresh_token | string | New refresh token (7 day validity, old one revoked) |
| expires_at | string | ISO 8601 timestamp when access token expires |
| expires_in | integer | Seconds until access token expires |

#### Error Responses

**400 Bad Request** - Invalid request format
```json
{
  "code": 400,
  "message": "Invalid request format",
  "error": "validation error details"
}
```

**401 Unauthorized** - Invalid or expired refresh token
```json
{
  "code": 401,
  "message": "Invalid or expired refresh token",
  "error": "token validation failed"
}
```

**401 Unauthorized** - Refresh token revoked
```json
{
  "code": 401,
  "message": "Refresh token has been revoked",
  "error": null
}
```

**401 Unauthorized** - User not found or disabled
```json
{
  "code": 401,
  "message": "User not found or disabled",
  "error": null
}
```

**500 Internal Server Error** - Server error
```json
{
  "code": 500,
  "message": "Failed to generate new token pair",
  "error": "internal error details"
}
```

## Security Features

### 1. Token Rotation

Each time a refresh token is used, a new token pair is generated:

- Old refresh token is immediately blacklisted
- Old refresh token is removed from active storage
- New refresh token has a fresh 7-day expiration
- Prevents token reuse attacks

### 2. Refresh Token Blacklist

- Used refresh tokens are added to Redis blacklist
- Blacklist entry expires when the token would naturally expire
- Prevents replay attacks if token is intercepted

### 3. Redis Storage

Refresh tokens are stored in Redis with the following keys:

```
refresh_token:{jti}           -> user_id (7 day TTL)
blacklist:refresh:{jti}       -> "1" (remaining TTL)
```

### 4. User Validation

Every refresh operation validates:

- Refresh token signature and expiration
- Token is not blacklisted
- Token exists in Redis storage
- Token belongs to the claimed user
- User account still exists and is active

## Login Flow Changes

The login endpoint now returns both access and refresh tokens:

### POST /api/v1/auth/login

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "jti": "550e8400-e29b-41d4-a716-446655440000",
    "expires_at": "2025-11-06T10:30:00Z",
    "expires_in": 7200,
    "user": {
      "id": "user-123",
      "username": "john.doe",
      "email": "john@example.com",
      "roles": [...]
    }
  }
}
```

## Client Implementation Guide

### Example JavaScript/TypeScript Client

```typescript
class AuthClient {
  private accessToken: string | null = null;
  private refreshToken: string | null = null;
  private tokenExpiresAt: Date | null = null;

  async login(username: string, password: string) {
    const response = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });

    const data = await response.json();

    this.accessToken = data.data.token;
    this.refreshToken = data.data.refresh_token;
    this.tokenExpiresAt = new Date(data.data.expires_at);

    // Store tokens securely (e.g., httpOnly cookies or secure storage)
    localStorage.setItem('refresh_token', this.refreshToken);

    return data;
  }

  async refreshAccessToken() {
    if (!this.refreshToken) {
      throw new Error('No refresh token available');
    }

    const response = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: this.refreshToken })
    });

    if (!response.ok) {
      // Refresh token invalid or expired - need to re-login
      this.logout();
      throw new Error('Token refresh failed');
    }

    const data = await response.json();

    this.accessToken = data.data.access_token;
    this.refreshToken = data.data.refresh_token; // New refresh token
    this.tokenExpiresAt = new Date(data.data.expires_at);

    // Update stored refresh token
    localStorage.setItem('refresh_token', this.refreshToken);

    return data;
  }

  async getAccessToken(): Promise<string> {
    // Check if token needs refresh (refresh 5 minutes before expiry)
    const now = new Date();
    const refreshThreshold = new Date(this.tokenExpiresAt.getTime() - 5 * 60 * 1000);

    if (now > refreshThreshold) {
      await this.refreshAccessToken();
    }

    return this.accessToken;
  }

  async makeAuthenticatedRequest(url: string, options: RequestInit = {}) {
    const token = await this.getAccessToken();

    return fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        'Authorization': `Bearer ${token}`
      }
    });
  }

  logout() {
    this.accessToken = null;
    this.refreshToken = null;
    this.tokenExpiresAt = null;
    localStorage.removeItem('refresh_token');
  }
}
```

### Example cURL Commands

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password123"
  }'
```

**Refresh Token:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**Use Access Token:**
```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## Best Practices

### For Clients

1. **Store Refresh Tokens Securely**
   - Use httpOnly cookies in web apps
   - Use secure storage (Keychain/Keystore) in mobile apps
   - Never store in localStorage for production web apps

2. **Automatic Token Refresh**
   - Implement automatic refresh before token expires
   - Handle 401 errors by attempting refresh
   - Redirect to login if refresh fails

3. **Handle Token Rotation**
   - Always use the new refresh token returned
   - Discard old refresh token immediately

4. **Error Handling**
   - Gracefully handle refresh failures
   - Clear tokens and redirect to login on 401

### For Backend

1. **Redis Configuration**
   - Ensure Redis persistence is enabled
   - Monitor Redis memory usage
   - Set appropriate TTL values

2. **Token Expiration**
   - Access token: 2 hours (configurable via JWT_EXPIRES_HOURS)
   - Refresh token: 7 days (hardcoded for security)

3. **Security**
   - Use strong JWT secrets (minimum 32 characters)
   - Enable HTTPS in production
   - Monitor for suspicious refresh patterns

## Configuration

### Environment Variables

```bash
# JWT Configuration
JWT_SECRET="your-super-secret-key-at-least-32-chars"
JWT_EXPIRES_HOURS=2

# Redis Configuration
REDIS_ADDR="localhost:6379"
REDIS_PASSWORD=""
REDIS_DB=0
REDIS_POOL_SIZE=10
```

### Database Schema

No database changes required. All refresh token data is stored in Redis.

## Troubleshooting

### Common Issues

**Issue**: "Refresh token not found or expired"
- **Cause**: Redis entry expired or was never created
- **Solution**: User must re-login

**Issue**: "Refresh token has been revoked"
- **Cause**: Token was already used (replay attack attempt)
- **Solution**: User must re-login

**Issue**: "User not found or disabled"
- **Cause**: User account was disabled or deleted
- **Solution**: Account needs to be re-enabled by admin

**Issue**: Token refresh succeeds but subsequent requests fail
- **Cause**: Client using old access token
- **Solution**: Ensure client updates access token after refresh

## Performance Considerations

- Each refresh operation requires 5-6 Redis operations
- Redis operations are pipelined where possible
- Expected latency: < 50ms for refresh operation
- Recommended: Use Redis Cluster for high-traffic scenarios

## Migration Notes

### Breaking Changes

The `LoginResponse` structure has changed:

**Before:**
```json
{
  "token": "...",
  "jti": "...",
  "expires_at": "...",
  "user": {...}
}
```

**After:**
```json
{
  "token": "...",
  "refresh_token": "...",  // NEW
  "jti": "...",
  "expires_at": "...",
  "expires_in": 7200,      // NEW
  "user": {...}
}
```

### Backward Compatibility

- Existing access tokens continue to work
- Old clients can ignore `refresh_token` field
- New clients should use refresh mechanism

## Testing

Run tests:
```bash
go test -v ./internal/auth/jwt/...
go test -v ./internal/auth/service/...
go test -v ./internal/auth/handler/...
```

## References

- [RFC 6749 - OAuth 2.0 Refresh Tokens](https://tools.ietf.org/html/rfc6749#section-1.5)
- [RFC 7519 - JSON Web Token (JWT)](https://tools.ietf.org/html/rfc7519)
- [OWASP Token Storage Best Practices](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
