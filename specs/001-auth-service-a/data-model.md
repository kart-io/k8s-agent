# Data Model: Auth Service

**Feature**: Authentication and Authorization Service
**Date**: 2025-10-09
**Database**: PostgreSQL
**ORM**: Database/SQL with lib/pq driver

## Overview

The auth-service implements a hierarchical RBAC (Role-Based Access Control) model with support for users, roles, and three types of permissions (menu, button, API). Additionally, it supports API Key authentication for service-to-service communication.

## Entity Relationship Diagram

```
┌─────────────┐       ┌──────────────┐       ┌────────────────┐
│    User     │──────>│  UserRole    │<──────│     Role       │
└─────────────┘       └──────────────┘       └────────────────┘
      │                                              │
      │                                              │
      │                                              v
      │                                       ┌──────────────┐
      │                                       │RolePermission│
      │                                       └──────────────┘
      │                                              │
      │                                              │
      v                                              v
┌─────────────┐                              ┌────────────────┐
│   APIKey    │                              │   Permission   │
└─────────────┘                              └────────────────┘
                                                     │
                                                     │ (self-ref)
                                                     │ parent_id
                                                     v
```

## Core Entities

### 1. User

Represents system users who can authenticate and access protected resources.

**Table**: `users`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | VARCHAR(36) | PRIMARY KEY | UUID identifier |
| username | VARCHAR(50) | UNIQUE, NOT NULL | Login username |
| password | VARCHAR(255) | NOT NULL | bcrypt hashed password |
| email | VARCHAR(100) | UNIQUE | User email address |
| real_name | VARCHAR(50) | | Full name |
| phone | VARCHAR(20) | | Phone number |
| avatar | VARCHAR(255) | | Avatar image URL |
| status | INT | DEFAULT 1 | 1=active, 0=inactive |
| created_at | TIMESTAMP | | Creation timestamp |
| updated_at | TIMESTAMP | | Last update timestamp |

**Indexes**:
- PRIMARY KEY on `id`
- UNIQUE INDEX on `username`
- UNIQUE INDEX on `email`
- INDEX on `status` (for filtering active users)

**Validation Rules**:
- `username`: 3-50 characters, alphanumeric and underscore only
- `password`: Minimum 8 characters (enforced at application level before hashing)
- `email`: Valid email format (RFC 5322)
- `phone`: Optional, E.164 format recommended
- `status`: 0 (inactive) or 1 (active)

**Business Rules**:
- Password must be hashed with bcrypt before storage
- Username and email must be unique across the system
- Inactive users (status=0) cannot authenticate
- Deletion is soft delete (set status=0) to maintain audit trail

**State Transitions**:
```
[New] --register--> [Active (status=1)]
       <--activate--
[Active] --deactivate--> [Inactive (status=0)]
```

### 2. Role

Represents user roles in the RBAC system.

**Table**: `roles`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | VARCHAR(36) | PRIMARY KEY | UUID identifier |
| name | VARCHAR(50) | UNIQUE, NOT NULL | Display name |
| code | VARCHAR(50) | UNIQUE, NOT NULL | System code (e.g., "super_admin") |
| description | TEXT | | Role description |
| status | INT | DEFAULT 1 | 1=active, 0=inactive |
| sort | INT | DEFAULT 0 | Display sort order |
| created_at | TIMESTAMP | | Creation timestamp |
| updated_at | TIMESTAMP | | Last update timestamp |

**Indexes**:
- PRIMARY KEY on `id`
- UNIQUE INDEX on `name`
- UNIQUE INDEX on `code`
- INDEX on `sort` (for ordered retrieval)

**Validation Rules**:
- `name`: 2-50 characters, human-readable
- `code`: Snake_case format, alphanumeric and underscore
- `status`: 0 (inactive) or 1 (active)
- `sort`: Non-negative integer

**Business Rules**:
- System roles (super_admin, admin, user) are protected from deletion
- Role code cannot be changed after creation (immutable)
- Deleting a role requires reassigning users to another role
- Inactive roles cannot be assigned to new users

**Default Roles**:
1. **super_admin** - Full system access
2. **admin** - Administrative access
3. **user** - Standard user access

### 3. Permission

Represents permissions in the system with hierarchical structure.

**Table**: `permissions`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | VARCHAR(36) | PRIMARY KEY | UUID identifier |
| parent_id | VARCHAR(36) | FOREIGN KEY → permissions(id) | Parent permission for hierarchy |
| name | VARCHAR(100) | NOT NULL | Display name |
| code | VARCHAR(100) | UNIQUE, NOT NULL | System code (e.g., "user:create") |
| type | VARCHAR(20) | NOT NULL | "menu", "button", or "api" |
| path | VARCHAR(255) | | URL path (for menu/api types) |
| method | VARCHAR(10) | | HTTP method (for api type) |
| component | VARCHAR(255) | | Frontend component (for menu type) |
| icon | VARCHAR(50) | | Icon name (for menu type) |
| sort | INT | DEFAULT 0 | Display sort order |
| status | INT | DEFAULT 1 | 1=active, 0=inactive |
| description | TEXT | | Permission description |
| created_at | TIMESTAMP | | Creation timestamp |
| updated_at | TIMESTAMP | | Last update timestamp |

**Indexes**:
- PRIMARY KEY on `id`
- UNIQUE INDEX on `code`
- INDEX on `parent_id` (for tree traversal)
- INDEX on `type` (for filtering by permission type)
- INDEX on `status, sort` (for ordered retrieval)

**Validation Rules**:
- `name`: 2-100 characters
- `code`: Format depends on type:
  - Menu: "module:menu" (e.g., "user:menu")
  - Button: "module:action" (e.g., "user:create")
  - API: "module:action" (e.g., "user:create")
- `type`: Must be "menu", "button", or "api"
- `method`: Only for api type, values: GET, POST, PUT, DELETE, PATCH
- `path`: Required for menu and api types

**Permission Types**:

1. **Menu Permission** (`type="menu"`):
   - Controls frontend navigation menu visibility
   - Required fields: `name`, `code`, `path`, `component`, `icon`
   - Tree structure via `parent_id`
   - Example: Dashboard menu item

2. **Button Permission** (`type="button"`):
   - Controls button/action visibility in UI
   - Required fields: `name`, `code`
   - Typically child of menu permission
   - Example: "Create User" button

3. **API Permission** (`type="api"`):
   - Controls backend API endpoint access
   - Required fields: `name`, `code`, `path`, `method`
   - Maps to actual HTTP endpoints
   - Example: POST /api/v1/users

**Business Rules**:
- Root permissions have `parent_id = NULL`
- Hierarchical structure maximum depth: 3 levels
- Permission code format: `resource:action` (colon-separated)
- API permissions must match actual endpoints
- Circular references in hierarchy are forbidden

**Permission Tree Example**:
```
System Management (menu)
├── User Management (menu)
│   ├── View Users (api: GET /api/v1/users)
│   ├── Create User (button + api: POST /api/v1/users)
│   ├── Edit User (button + api: PUT /api/v1/users/:id)
│   └── Delete User (button + api: DELETE /api/v1/users/:id)
└── Role Management (menu)
    ├── View Roles (api: GET /api/v1/roles)
    └── ...
```

### 4. UserRole (Junction Table)

Maps users to roles (many-to-many relationship).

**Table**: `user_roles`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| user_id | VARCHAR(36) | FOREIGN KEY → users(id) | User identifier |
| role_id | VARCHAR(36) | FOREIGN KEY → roles(id) | Role identifier |

**Indexes**:
- PRIMARY KEY on `(user_id, role_id)`
- INDEX on `role_id` (for reverse lookups)

**Business Rules**:
- One user can have multiple roles
- One role can be assigned to multiple users
- Deleting a user cascades to delete user_roles entries
- Deleting a role requires handling user_roles entries
- At least one role should be assigned to active users

**Constraints**:
- CASCADE DELETE on user_id
- RESTRICT DELETE on role_id (must reassign users first)

### 5. RolePermission (Junction Table)

Maps roles to permissions (many-to-many relationship).

**Table**: `role_permissions`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| role_id | VARCHAR(36) | FOREIGN KEY → roles(id) | Role identifier |
| permission_id | VARCHAR(36) | FOREIGN KEY → permissions(id) | Permission identifier |

**Indexes**:
- PRIMARY KEY on `(role_id, permission_id)`
- INDEX on `permission_id` (for reverse lookups)

**Business Rules**:
- One role can have multiple permissions
- One permission can be assigned to multiple roles
- Super admin role gets all permissions by default
- Assigning parent permission auto-includes children (at query time)
- Deleting a permission cascades to delete role_permissions entries

**Constraints**:
- CASCADE DELETE on both foreign keys

### 6. APIKey

Represents API keys for service-to-service authentication.

**Table**: `api_keys`

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | VARCHAR(36) | PRIMARY KEY | UUID identifier |
| name | VARCHAR(100) | NOT NULL | Friendly name |
| key | VARCHAR(255) | UNIQUE, NOT NULL | Public API key |
| secret | VARCHAR(255) | NOT NULL | Secret key (hashed) |
| user_id | VARCHAR(36) | FOREIGN KEY → users(id) | Owner user |
| description | TEXT | | Key description/purpose |
| expires_at | TIMESTAMP | | Expiration timestamp (NULL = never) |
| status | INT | DEFAULT 1 | 1=active, 0=inactive |
| last_used_at | TIMESTAMP | | Last usage timestamp |
| created_at | TIMESTAMP | | Creation timestamp |
| updated_at | TIMESTAMP | | Last update timestamp |

**Indexes**:
- PRIMARY KEY on `id`
- UNIQUE INDEX on `key`
- INDEX on `user_id` (for user's key lookup)
- INDEX on `expires_at, status` (for cleanup queries)

**Validation Rules**:
- `name`: 3-100 characters
- `key`: Auto-generated, format: `ak_` + 32 random characters
- `secret`: Auto-generated, hashed with bcrypt before storage
- `expires_at`: Optional, must be future date
- `status`: 0 (inactive) or 1 (active)

**Business Rules**:
- Key and secret generated server-side (cryptographically secure)
- Secret shown only once at creation, then hashed
- Expired keys (expires_at < now) cannot authenticate
- Inactive keys cannot authenticate
- Usage tracked via last_used_at for auditing
- User deletion should cascade to API keys or set NULL

**State Transitions**:
```
[New] --create--> [Active (status=1)]
       <--activate--
[Active] --deactivate--> [Inactive (status=0)]
[Active] --expire--> [Expired (expires_at < now)]
[Active/Inactive/Expired] --delete--> [Deleted]
```

## Database Initialization

### Migration Strategy

1. **Initial Schema Creation**:
   - Create tables in dependency order
   - Add indexes and constraints
   - Set default values

2. **Seed Data**:
   - Create super admin user (username: admin, password: admin123)
   - Create default roles (super_admin, admin, user)
   - Create default permissions for system management
   - Assign all permissions to super_admin role
   - Assign super_admin role to admin user

### Sample Seed Data

**Default Super Admin User**:
```sql
INSERT INTO users (id, username, password, email, real_name, status, created_at, updated_at)
VALUES (
  'uuid-admin',
  'admin',
  '$2a$10$...', -- bcrypt hash of 'admin123'
  'admin@example.com',
  'System Administrator',
  1,
  NOW(),
  NOW()
);
```

**Default Roles**:
```sql
INSERT INTO roles (id, name, code, description, status, sort) VALUES
('uuid-super-admin', '超级管理员', 'super_admin', 'Full system access', 1, 1),
('uuid-admin', '管理员', 'admin', 'Administrative access', 1, 2),
('uuid-user', '普通用户', 'user', 'Standard user access', 1, 3);
```

**Default Permissions** (Examples):
```sql
-- System management menu
INSERT INTO permissions (id, parent_id, name, code, type, path, component, icon, sort, status)
VALUES ('uuid-system', NULL, 'System Management', 'system:menu', 'menu', '/system', 'SystemLayout', 'setting', 1, 1);

-- User management submenu
INSERT INTO permissions (id, parent_id, name, code, type, path, component, icon, sort, status)
VALUES ('uuid-user-menu', 'uuid-system', 'User Management', 'user:menu', 'menu', '/system/users', 'UserManagement', 'user', 1, 1);

-- API permissions for user CRUD
INSERT INTO permissions (id, parent_id, name, code, type, path, method, sort, status) VALUES
('uuid-user-list', 'uuid-user-menu', 'View Users', 'user:list', 'api', '/api/v1/users', 'GET', 1, 1),
('uuid-user-create', 'uuid-user-menu', 'Create User', 'user:create', 'api', '/api/v1/users', 'POST', 2, 1),
('uuid-user-update', 'uuid-user-menu', 'Update User', 'user:update', 'api', '/api/v1/users/:id', 'PUT', 3, 1),
('uuid-user-delete', 'uuid-user-menu', 'Delete User', 'user:delete', 'api', '/api/v1/users/:id', 'DELETE', 4, 1);
```

## Query Patterns

### Common Queries

**1. User Authentication**:
```sql
SELECT id, username, password, email, real_name, status
FROM users
WHERE username = $1 AND status = 1;
```

**2. Get User Roles**:
```sql
SELECT r.id, r.name, r.code
FROM roles r
JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1 AND r.status = 1;
```

**3. Get User Permissions**:
```sql
SELECT DISTINCT p.id, p.code, p.type, p.path, p.method
FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = $1 AND p.status = 1;
```

**4. Get User Menu Tree**:
```sql
WITH RECURSIVE menu_tree AS (
  -- Root level menus
  SELECT p.id, p.parent_id, p.name, p.code, p.path, p.component, p.icon, p.sort
  FROM permissions p
  JOIN role_permissions rp ON p.id = rp.permission_id
  JOIN user_roles ur ON rp.role_id = ur.role_id
  WHERE ur.user_id = $1
    AND p.type = 'menu'
    AND p.parent_id IS NULL
    AND p.status = 1

  UNION ALL

  -- Child menus
  SELECT p.id, p.parent_id, p.name, p.code, p.path, p.component, p.icon, p.sort
  FROM permissions p
  JOIN role_permissions rp ON p.id = rp.permission_id
  JOIN user_roles ur ON rp.role_id = ur.role_id
  JOIN menu_tree mt ON p.parent_id = mt.id
  WHERE ur.user_id = $1
    AND p.type = 'menu'
    AND p.status = 1
)
SELECT * FROM menu_tree ORDER BY sort;
```

**5. Check API Permission**:
```sql
SELECT EXISTS (
  SELECT 1
  FROM permissions p
  JOIN role_permissions rp ON p.id = rp.permission_id
  JOIN user_roles ur ON rp.role_id = ur.role_id
  WHERE ur.user_id = $1
    AND p.type = 'api'
    AND p.path = $2
    AND p.method = $3
    AND p.status = 1
) AS has_permission;
```

**6. Validate API Key**:
```sql
SELECT id, key, secret, user_id, status, expires_at
FROM api_keys
WHERE key = $1
  AND status = 1
  AND (expires_at IS NULL OR expires_at > NOW());
```

## Performance Optimization

### Indexing Strategy

- **Users**: Index on username, email for login lookups
- **Roles**: Index on code for role-based checks
- **Permissions**: Index on code, type, path+method for permission checks
- **UserRoles**: Composite index on (user_id, role_id) for fast joins
- **RolePermissions**: Composite index on (role_id, permission_id) for fast joins
- **APIKeys**: Index on key for authentication, expires_at for cleanup

### Caching Recommendations

- **User permissions**: Cache in Redis with TTL (e.g., 15 minutes)
- **Permission tree**: Cache entire tree for menu rendering
- **Role permissions**: Cache role→permissions mapping
- **API key validation**: Cache valid keys with TTL

### Data Archival

- **Inactive users**: Archive after 1 year of inactivity
- **Expired API keys**: Archive after 90 days past expiration
- **Audit logs**: Retain for compliance period, then archive

## Validation and Constraints Summary

| Entity | Unique Constraints | Check Constraints | Foreign Keys |
|--------|-------------------|------------------|--------------|
| User | username, email | status IN (0,1) | - |
| Role | name, code | status IN (0,1) | - |
| Permission | code | status IN (0,1), type IN ('menu','button','api') | parent_id → permissions(id) |
| UserRole | (user_id, role_id) | - | user_id → users(id), role_id → roles(id) |
| RolePermission | (role_id, permission_id) | - | role_id → roles(id), permission_id → permissions(id) |
| APIKey | key | status IN (0,1) | user_id → users(id) |

## Migration Scripts

Migration scripts will be maintained in `/auth-service/internal/storage/migrate.go` with the following functions:

- `AutoMigrate()`: Create tables and indexes
- `Seed()`: Insert default data
- `Rollback()`: Undo migrations (for development)
- `Version()`: Track migration version

## Future Enhancements

1. **Multi-tenancy**:
   - Add `tenant_id` column to all tables
   - Partition data by tenant
   - Tenant-specific roles and permissions

2. **Permission Templates**:
   - Pre-defined permission sets
   - Quick role creation from templates

3. **Temporal Data**:
   - Track permission changes over time
   - Role assignment history
   - Audit trail for compliance

4. **Soft Delete**:
   - Add `deleted_at` column
   - Implement soft delete for all entities
   - Scheduled cleanup of soft-deleted records
