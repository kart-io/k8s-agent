# API Contract Preservation

**Feature**: GORM and kart-io/logger Integration
**Purpose**: Document existing API contracts that MUST be preserved during refactoring

## Overview

This refactoring MUST NOT change any existing API endpoints, request formats, or response formats. All endpoints listed here must continue to function identically after GORM and kart-io/logger integration.

## Contract Validation

**Validation Method**: API integration tests comparing responses before/after migration

**Success Criteria**: SC-009 - All existing API endpoints return identical responses

---

## Authentication Endpoints

### POST /api/v1/auth/login

**Request**:
```json
{
  "username": "string (required)",
  "password": "string (required)"
}
```

**Success Response** (200 OK):
```json
{
  "token": "string (JWT)",
  "expires_at": "2024-01-01T00:00:00Z",
  "user": {
    "id": "string (UUID)",
    "username": "string",
    "email": "string",
    "real_name": "string",
    "avatar": "string",
    "roles": [
      {
        "id": "string",
        "name": "string",
        "code": "string"
      }
    ]
  }
}
```

**Error Responses**:
- 400: Invalid request format
- 401: Invalid credentials

---

### POST /api/v1/auth/logout

**Headers**: `Authorization: Bearer <token>`

**Success Response** (200 OK):
```json
{
  "message": "Logged out successfully"
}
```

---

### GET /api/v1/auth/me

**Headers**: `Authorization: Bearer <token>`

**Success Response** (200 OK):
```json
{
  "id": "string",
  "username": "string",
  "email": "string",
  "real_name": "string",
  "avatar": "string",
  "roles": [...]
}
```

---

### GET /api/v1/auth/menus

**Headers**: `Authorization: Bearer <token>`

**Success Response** (200 OK):
```json
[
  {
    "id": "string",
    "parent_id": "string",
    "name": "string",
    "path": "string",
    "component": "string",
    "icon": "string",
    "sort": 0,
    "children": [...]
  }
]
```

---

### POST /api/v1/auth/check

**Request**:
```json
{
  "user_id": "string",
  "path": "string",
  "method": "string"
}
```

**Success Response** (200 OK):
```json
{
  "allowed": true
}
```

---

## User Management Endpoints

### GET /api/v1/users

**Headers**: `Authorization: Bearer <token>`

**Query Parameters**:
- `page`: integer (default: 1)
- `page_size`: integer (default: 20, max: 100)
- `username`: string (filter)
- `email`: string (filter)
- `status`: integer (0 or 1)

**Success Response** (200 OK):
```json
{
  "items": [...],
  "total": 100,
  "page": 1,
  "page_size": 20,
  "total_pages": 5
}
```

---

### GET /api/v1/users/:id

**Headers**: `Authorization: Bearer <token>`

**Success Response** (200 OK):
```json
{
  "id": "string",
  "username": "string",
  "email": "string",
  "real_name": "string",
  "phone": "string",
  "avatar": "string",
  "status": 1,
  "role_ids": ["uuid1", "uuid2"],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

### POST /api/v1/users

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "username": "string (required, 3-50 chars)",
  "password": "string (required, min 8 chars)",
  "email": "string (required, valid email)",
  "real_name": "string",
  "phone": "string",
  "avatar": "string",
  "role_ids": ["uuid1", "uuid2"]
}
```

**Success Response** (201 Created):
```json
{
  "id": "string",
  "username": "string",
  ...
}
```

---

### PUT /api/v1/users/:id

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "email": "string",
  "real_name": "string",
  "phone": "string",
  "avatar": "string",
  "status": 0 or 1,
  "role_ids": ["uuid1", "uuid2"]
}
```

**Success Response** (200 OK):
```json
{
  "id": "string",
  "username": "string",
  ...
}
```

---

### DELETE /api/v1/users/:id

**Headers**: `Authorization: Bearer <token>`

**Success Response** (200 OK):
```json
{
  "message": "User deleted successfully"
}
```

---

### POST /api/v1/users/:id/roles

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "role_ids": ["uuid1", "uuid2"]
}
```

**Success Response** (200 OK):
```json
{
  "message": "Roles assigned successfully"
}
```

---

## Role Management Endpoints

### GET /api/v1/roles

**Headers**: `Authorization: Bearer <token>`

**Query Parameters**: page, page_size, name, code, status

**Success Response** (200 OK): Paginated list of roles

---

### GET /api/v1/roles/:id

**Headers**: `Authorization: Bearer <token>`

**Success Response** (200 OK): Role details

---

### POST /api/v1/roles

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "name": "string (required, 2-50 chars)",
  "code": "string (required)",
  "description": "string",
  "status": 0 or 1,
  "sort": 0
}
```

---

### PUT /api/v1/roles/:id

**Headers**: `Authorization: Bearer <token>`

**Request**: Same as POST

---

### DELETE /api/v1/roles/:id

**Headers**: `Authorization: Bearer <token>`

**Note**: System roles (super_admin, admin, user) cannot be deleted

---

### POST /api/v1/roles/:id/permissions

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "permission_ids": ["uuid1", "uuid2"]
}
```

---

### GET /api/v1/roles/:id/permissions

**Headers**: `Authorization: Bearer <token>`

**Success Response** (200 OK): List of permissions for the role

---

## Permission Management Endpoints

### GET /api/v1/permissions

**Headers**: `Authorization: Bearer <token>`

**Query Parameters**: page, page_size, name, code, type, status

---

### GET /api/v1/permissions/tree

**Headers**: `Authorization: Bearer <token>`

**Success Response** (200 OK): Hierarchical permission tree

---

### GET /api/v1/permissions/:id

**Headers**: `Authorization: Bearer <token>`

---

### POST /api/v1/permissions

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "parent_id": "string",
  "name": "string (required, 2-100 chars)",
  "code": "string (required)",
  "type": "menu|button|api",
  "path": "string",
  "method": "string",
  "component": "string",
  "icon": "string",
  "sort": 0,
  "status": 0 or 1,
  "description": "string"
}
```

---

### PUT /api/v1/permissions/:id

**Headers**: `Authorization: Bearer <token>`

---

### DELETE /api/v1/permissions/:id

**Headers**: `Authorization: Bearer <token>`

---

## API Key Management Endpoints

### GET /api/v1/api-keys

**Headers**: `Authorization: Bearer <token>`

**Success Response** (200 OK):
```json
{
  "items": [
    {
      "id": "string",
      "name": "string",
      "key": "ak_xxxxx",
      "secret": "****", // Masked
      "description": "string",
      "expires_at": "2024-01-01T00:00:00Z",
      "status": 1,
      "last_used_at": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### POST /api/v1/api-keys

**Headers**: `Authorization: Bearer <token>`

**Request**:
```json
{
  "name": "string (required, 3-100 chars)",
  "description": "string",
  "expires_at": "2024-01-01T00:00:00Z"
}
```

**Success Response** (201 Created):
```json
{
  "id": "string",
  "name": "string",
  "key": "ak_xxxxx",
  "secret": "sk_xxxxx", // Plain text, shown only once!
  "description": "string",
  "expires_at": "2024-01-01T00:00:00Z",
  "created_at": "2024-01-01T00:00:00Z",
  "warning": "Save the secret now. You won't be able to see it again!"
}
```

---

### DELETE /api/v1/api-keys/:id

**Headers**: `Authorization: Bearer <token>`

**Success Response** (200 OK):
```json
{
  "message": "API key deleted successfully"
}
```

---

## Health and Monitoring Endpoints

### GET /health

**No authentication required**

**Success Response** (200 OK):
```json
{
  "status": "ok",
  "time": "2024-01-01T00:00:00Z",
  "checks": {
    "database": "ok",
    "redis": "ok"
  }
}
```

**Degraded Response** (503 Service Unavailable):
```json
{
  "status": "degraded",
  "time": "2024-01-01T00:00:00Z",
  "checks": {
    "database": "error",
    "redis": "ok"
  }
}
```

---

### GET /metrics

**No authentication required**

**Success Response** (200 OK): Prometheus metrics in text format

---

## Error Response Format

**Standard Error Format** (all error responses):
```json
{
  "error": "Error message",
  "code": 400,
  "details": "Detailed error description"
}
```

**Error Codes**:
- 400: Bad Request / Validation Error
- 401: Unauthorized / Authentication Failed
- 403: Forbidden / Permission Denied
- 404: Not Found
- 409: Conflict (e.g., duplicate username)
- 500: Internal Server Error
- 503: Service Unavailable

---

## Contract Test Requirements

### Test Coverage

**Each endpoint must have**:
1. Happy path test (valid request → expected response)
2. Validation test (invalid request → 400 error)
3. Authentication test (missing/invalid token → 401 error)
4. Permission test (insufficient permissions → 403 error)

### Test Implementation

**Before GORM Migration**:
```bash
# Capture baseline responses
./test/capture-baseline-responses.sh > baseline.json
```

**After GORM Migration**:
```bash
# Compare responses
./test/compare-api-responses.sh baseline.json current.json
```

**Test Script Example**:
```bash
#!/bin/bash
# Test login endpoint
response=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

# Verify response structure
echo "$response" | jq -e '.token' > /dev/null || exit 1
echo "$response" | jq -e '.user.id' > /dev/null || exit 1
echo "$response" | jq -e '.user.roles' > /dev/null || exit 1
```

---

## Migration Validation Checklist

- [ ] All authentication endpoints return identical responses
- [ ] All user management endpoints return identical responses
- [ ] All role management endpoints return identical responses
- [ ] All permission management endpoints return identical responses
- [ ] All API key management endpoints return identical responses
- [ ] Health check endpoint format preserved
- [ ] Metrics endpoint format preserved
- [ ] Error response format consistent
- [ ] Pagination format consistent
- [ ] JWT token validation unchanged
- [ ] API key authentication unchanged
- [ ] Permission checks unchanged
