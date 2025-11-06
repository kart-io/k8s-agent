# Token Refresh Quick Reference

## For Frontend Developers

### Login Flow
```javascript
// 1. Login and get tokens
const response = await fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ username, password })
});

const { token, refresh_token, expires_in } = await response.json().data;

// 2. Store both tokens securely
localStorage.setItem('access_token', token);
localStorage.setItem('refresh_token', refresh_token);
```

### Using Access Token
```javascript
// Use access token in Authorization header
const response = await fetch('/api/v1/users', {
  headers: { 'Authorization': `Bearer ${accessToken}` }
});
```

### Refresh Flow
```javascript
// When access token expires (401 error), refresh it
const response = await fetch('/api/v1/auth/refresh', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ refresh_token: refreshToken })
});

if (response.ok) {
  const { access_token, refresh_token } = await response.json().data;
  // Update stored tokens
  localStorage.setItem('access_token', access_token);
  localStorage.setItem('refresh_token', refresh_token);
} else {
  // Refresh failed - redirect to login
  window.location.href = '/login';
}
```

### Automatic Refresh (Recommended)
```javascript
// Refresh 5 minutes before expiration
const shouldRefresh = (expiresAt) => {
  return new Date(expiresAt) - new Date() < 5 * 60 * 1000;
};

// Wrap API calls with automatic refresh
async function apiCall(url, options = {}) {
  if (shouldRefresh(tokenExpiresAt)) {
    await refreshToken();
  }

  const token = localStorage.getItem('access_token');
  return fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${token}`
    }
  });
}
```

## For Backend Developers

### Testing with cURL

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password123"}'
```

**Refresh:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"YOUR_REFRESH_TOKEN"}'
```

**Use Access Token:**
```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Redis Keys

```
refresh_token:{jti}           → user_id (7d TTL)
blacklist:refresh:{jti}       → "1" (remaining TTL)
session:{jti}                 → session_data (2h TTL)
user:sessions:{user_id}       → sorted set of JTIs
revoked:{jti}                 → revoked_data (2h TTL)
```

### Configuration

```bash
# .env or environment variables
JWT_SECRET=your-super-secret-key
JWT_EXPIRES_HOURS=2
REDIS_ADDR=localhost:6379
```

## Token Lifetimes

| Token Type | Lifetime | Purpose |
|------------|----------|---------|
| Access Token | 2 hours | API authentication |
| Refresh Token | 7 days | Get new access tokens |

## Error Codes

| Code | Message | Action |
|------|---------|--------|
| 400 | Invalid request format | Check request body |
| 401 | Invalid/expired refresh token | User must re-login |
| 401 | Refresh token revoked | User must re-login |
| 401 | User not found/disabled | Contact admin |
| 500 | Server error | Check logs, retry |

## Security Best Practices

### Client Side
- Store refresh tokens in httpOnly cookies (web)
- Store in Keychain/Keystore (mobile)
- Never expose refresh tokens in URL/logs
- Implement automatic refresh before expiry
- Clear tokens on logout

### Server Side
- Use strong JWT secret (32+ characters)
- Enable HTTPS in production
- Monitor Redis memory usage
- Enable Redis persistence
- Rate limit refresh endpoint
- Log suspicious refresh patterns

## Common Issues

**"refresh token not found"**
→ Token expired or Redis cleared → User re-login

**"refresh token revoked"**
→ Token already used (replay attack) → User re-login

**401 after successful refresh**
→ Client using old access token → Update token storage

**High Redis memory**
→ Too many refresh tokens → Check TTL settings

## Code Examples

### Go Client
```go
type AuthClient struct {
    accessToken  string
    refreshToken string
    expiresAt    time.Time
}

func (c *AuthClient) GetAccessToken() (string, error) {
    if time.Now().Add(5*time.Minute).After(c.expiresAt) {
        if err := c.refresh(); err != nil {
            return "", err
        }
    }
    return c.accessToken, nil
}
```

### Python Client
```python
class AuthClient:
    def get_access_token(self):
        if self.should_refresh():
            self.refresh()
        return self.access_token

    def should_refresh(self):
        return datetime.now() + timedelta(minutes=5) > self.expires_at
```

## Monitoring

### Key Metrics
- Refresh rate per user
- Failed refresh attempts
- Token rotation success rate
- Redis hit rate
- Average refresh latency

### Alerts
- High failed refresh rate (> 5%)
- Unusual refresh patterns
- Redis connection failures
- Token blacklist size growing

## Files Reference

| File | Purpose |
|------|---------|
| `internal/auth/jwt/jwt.go` | Token generation |
| `internal/auth/handler/auth_handler.go` | HTTP handlers |
| `internal/auth/service/auth_service.go` | Business logic |
| `internal/auth/storage/redis.go` | Redis operations |
| `docs/auth/TOKEN_REFRESH_API.md` | Full API docs |

## Quick Debugging

```bash
# Check Redis keys
redis-cli
> KEYS refresh_token:*
> GET refresh_token:{jti}
> TTL refresh_token:{jti}

# Check if token is blacklisted
> EXISTS blacklist:refresh:{jti}

# Monitor refresh operations
> MONITOR

# Check active sessions
> ZRANGE user:sessions:{user_id} 0 -1
```

## Need Help?

- Full Documentation: `docs/auth/TOKEN_REFRESH_API.md`
- Implementation Details: `docs/auth/TOKEN_REFRESH_IMPLEMENTATION.md`
- JWT Package: `internal/auth/jwt/jwt.go`
- Tests: `internal/auth/jwt/jwt_test.go`
