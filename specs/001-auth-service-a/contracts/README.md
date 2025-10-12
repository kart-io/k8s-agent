# API Contracts Overview

This directory contains OpenAPI 3.0 specifications for the Auth Service REST API.

## Files

- **auth.yaml** - Authentication endpoints (login, logout, user info, menus)
- **users.yaml** - User management CRUD operations (to be generated)
- **roles.yaml** - Role management CRUD operations (to be generated)
- **permissions.yaml** - Permission management CRUD operations (to be generated)
- **api-keys.yaml** - API Key management operations (to be generated)

## Base URL

- Development: `http://localhost:8090/api/v1`
- Production: `https://api.example.com/api/v1`

## Authentication

All endpoints (except `/auth/login`) require JWT Bearer token authentication:

```
Authorization: Bearer <jwt_token>
```

## API Endpoint Summary

### Authentication (`/auth/*`)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auth/login` | User login | No |
| POST | `/auth/logout` | User logout | Yes |
| GET | `/auth/me` | Get current user info | Yes |
| GET | `/auth/menus` | Get user menu tree | Yes |
| POST | `/auth/check` | Check user permission | No |

### Users (`/users/*`)

| Method | Endpoint | Description | Permission |
|--------|----------|-------------|------------|
| GET | `/users` | List all users | `user:list` |
| GET | `/users/:id` | Get user by ID | `user:view` |
| POST | `/users` | Create new user | `user:create` |
| PUT | `/users/:id` | Update user | `user:update` |
| DELETE | `/users/:id` | Delete user | `user:delete` |
| POST | `/users/:id/roles` | Assign roles to user | `user:assign` |

### Roles (`/roles/*`)

| Method | Endpoint | Description | Permission |
|--------|----------|-------------|------------|
| GET | `/roles` | List all roles | `role:list` |
| GET | `/roles/:id` | Get role by ID | `role:view` |
| POST | `/roles` | Create new role | `role:create` |
| PUT | `/roles/:id` | Update role | `role:update` |
| DELETE | `/roles/:id` | Delete role | `role:delete` |
| POST | `/roles/:id/permissions` | Assign permissions | `role:assign` |

### Permissions (`/permissions/*`)

| Method | Endpoint | Description | Permission |
|--------|----------|-------------|------------|
| GET | `/permissions` | List all permissions | `permission:list` |
| GET | `/permissions/tree` | Get permission tree | `permission:list` |
| GET | `/permissions/:id` | Get permission by ID | `permission:view` |
| POST | `/permissions` | Create new permission | `permission:create` |
| PUT | `/permissions/:id` | Update permission | `permission:update` |
| DELETE | `/permissions/:id` | Delete permission | `permission:delete` |

### API Keys (`/api-keys/*`)

| Method | Endpoint | Description | Permission |
|--------|----------|-------------|------------|
| GET | `/api-keys` | List user's API keys | `apikey:list` |
| POST | `/api-keys` | Create new API key | `apikey:create` |
| DELETE | `/api-keys/:id` | Delete API key | `apikey:delete` |

## Common Response Codes

- `200` - Success
- `201` - Created
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (authentication required/failed)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `500` - Internal Server Error

## Error Response Format

All error responses follow this structure:

```json
{
  "error": "Error Type",
  "code": 400,
  "details": "Detailed error message"
}
```

## Pagination

List endpoints support pagination with query parameters:

- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 20, max: 100)
- `sort` - Sort field (default: created_at)
- `order` - Sort order: `asc` or `desc` (default: desc)

Example:
```
GET /api/v1/users?page=2&page_size=50&sort=username&order=asc
```

Paginated response format:
```json
{
  "items": [...],
  "total": 150,
  "page": 2,
  "page_size": 50,
  "total_pages": 3
}
```

## Filtering

List endpoints support filtering with query parameters matching entity fields:

```
GET /api/v1/users?status=1&role=admin
GET /api/v1/permissions?type=api&status=1
```

## Testing with Swagger UI

Once implemented, the Swagger UI will be available at:

```
http://localhost:8090/docs
```

## Generating Client SDKs

Use the OpenAPI Generator to create client SDKs:

```bash
# JavaScript/TypeScript client
openapi-generator generate -i auth.yaml -g typescript-axios -o clients/ts

# Python client
openapi-generator generate -i auth.yaml -g python -o clients/python

# Go client
openapi-generator generate -i auth.yaml -g go -o clients/go
```

## Validation

Validate OpenAPI specs using:

```bash
# Using swagger-cli
swagger-cli validate auth.yaml

# Using openapi-generator
openapi-generator validate -i auth.yaml
```

## Next Steps

1. Complete remaining OpenAPI specification files (users, roles, permissions, api-keys)
2. Integrate Swagger UI into the application
3. Set up API contract testing
4. Generate client SDKs for frontend integration
5. Document authentication flows with sequence diagrams
