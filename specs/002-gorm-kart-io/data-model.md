# Data Model: GORM and kart-io/logger Integration

**Feature**: Code Optimization - GORM and kart-io/logger Integration
**Branch**: `002-gorm-kart-io`
**Date**: 2025-10-09

## Overview

This document describes the GORM model definitions that will replace the current raw SQL implementation. All models map to existing database tables without requiring schema changes.

## Entity Relationships

```
User ────────< UserRole >──────── Role
                                   │
                                   │
                                   └──< RolePermission >──── Permission
                                                              │
                                                              └──── Permission (self-reference for hierarchy)

User ────────< APIKey
```

---

## Core Entities

### User

**Purpose**: Represents system users with authentication credentials and profile information

**Table Name**: `users`

**GORM Model Definition**:
```go
type User struct {
    ID        string    `gorm:"column:id;primaryKey;type:varchar(36)"`
    Username  string    `gorm:"column:username;uniqueIndex;not null;type:varchar(50)"`
    Password  string    `gorm:"column:password;not null;type:varchar(255)"` // bcrypt hash
    Email     string    `gorm:"column:email;uniqueIndex;type:varchar(100)"`
    RealName  string    `gorm:"column:real_name;type:varchar(50)"`
    Phone     string    `gorm:"column:phone;type:varchar(20)"`
    Avatar    string    `gorm:"column:avatar;type:varchar(255)"`
    Status    int       `gorm:"column:status;default:1;type:integer"` // 1=active, 0=disabled
    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`

    // Relationships
    Roles   []Role   `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:UserID;References:ID;joinReferences:RoleID"`
    APIKeys []APIKey `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

func (User) TableName() string {
    return "users"
}
```

**Fields**:
- `ID` (string, PK): UUID primary key
- `Username` (string, unique, required): Login username
- `Password` (string, required): bcrypt password hash
- `Email` (string, unique): User email address
- `RealName` (string): Display name
- `Phone` (string): Contact phone number
- `Avatar` (string): Profile picture URL
- `Status` (int): Account status (1=active, 0=disabled)
- `CreatedAt` (timestamp): Account creation time
- `UpdatedAt` (timestamp): Last modification time

**Validation Rules**:
- Username: 3-50 characters, alphanumeric + underscore + hyphen
- Password: Minimum 8 characters (enforced before hashing)
- Email: Valid email format
- Status: Must be 0 or 1

**Indexes**:
- Primary key on `id`
- Unique index on `username`
- Unique index on `email`
- Index on `status` (for active user queries)

---

### Role

**Purpose**: Represents user roles for RBAC system

**Table Name**: `roles`

**GORM Model Definition**:
```go
type Role struct {
    ID          string    `gorm:"column:id;primaryKey;type:varchar(36)"`
    Name        string    `gorm:"column:name;uniqueIndex;not null;type:varchar(50)"`
    Code        string    `gorm:"column:code;uniqueIndex;not null;type:varchar(50)"`
    Description string    `gorm:"column:description;type:text"`
    Status      int       `gorm:"column:status;default:1;type:integer"`
    Sort        int       `gorm:"column:sort;default:0;type:integer"`
    CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`

    // Relationships
    Users       []User       `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:RoleID;References:ID;joinReferences:UserID"`
    Permissions []Permission `gorm:"many2many:role_permissions;foreignKey:ID;joinForeignKey:RoleID;References:ID;joinReferences:PermissionID"`
}

func (Role) TableName() string {
    return "roles"
}
```

**Fields**:
- `ID` (string, PK): UUID primary key
- `Name` (string, unique, required): Display name (e.g., "Super Admin")
- `Code` (string, unique, required): System code (e.g., "super_admin")
- `Description` (string): Role description
- `Status` (int): Role status (1=active, 0=disabled)
- `Sort` (int): Display sort order
- `CreatedAt` (timestamp): Creation time
- `UpdatedAt` (timestamp): Last modification time

**Validation Rules**:
- Name: 2-50 characters
- Code: Required, used for permission checks
- Status: Must be 0 or 1

**Indexes**:
- Primary key on `id`
- Unique index on `name`
- Unique index on `code`
- Index on `sort` (for ordered queries)

**Business Rules**:
- System roles (super_admin, admin, user) cannot be deleted
- Role code cannot be changed after creation

---

### Permission

**Purpose**: Represents hierarchical system permissions (menu, button, API)

**Table Name**: `permissions`

**GORM Model Definition**:
```go
type Permission struct {
    ID          string    `gorm:"column:id;primaryKey;type:varchar(36)"`
    ParentID    string    `gorm:"column:parent_id;index;type:varchar(36)"`
    Name        string    `gorm:"column:name;not null;type:varchar(100)"`
    Code        string    `gorm:"column:code;uniqueIndex;not null;type:varchar(100)"`
    Type        string    `gorm:"column:type;not null;type:varchar(20)"` // menu, button, api
    Path        string    `gorm:"column:path;type:varchar(255)"`
    Method      string    `gorm:"column:method;type:varchar(10)"`        // For API type
    Component   string    `gorm:"column:component;type:varchar(255)"`    // For menu type
    Icon        string    `gorm:"column:icon;type:varchar(50)"`
    Sort        int       `gorm:"column:sort;default:0;type:integer"`
    Status      int       `gorm:"column:status;default:1;type:integer"`
    Description string    `gorm:"column:description;type:text"`
    CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`

    // Relationships
    Parent   *Permission   `gorm:"foreignKey:ParentID;references:ID"`
    Children []Permission  `gorm:"foreignKey:ParentID;references:ID"`
    Roles    []Role        `gorm:"many2many:role_permissions;foreignKey:ID;joinForeignKey:PermissionID;References:ID;joinReferences:RoleID"`
}

func (Permission) TableName() string {
    return "permissions"
}
```

**Fields**:
- `ID` (string, PK): UUID primary key
- `ParentID` (string, FK): Parent permission ID (for hierarchy)
- `Name` (string, required): Display name
- `Code` (string, unique, required): System code (e.g., "user:create")
- `Type` (string, required): Permission type (menu/button/api)
- `Path` (string): URL path for menu/API
- `Method` (string): HTTP method for API type
- `Component` (string): Component path for menu type
- `Icon` (string): Icon name for menu type
- `Sort` (int): Display sort order
- `Status` (int): Permission status (1=active, 0=disabled)
- `Description` (string): Permission description
- `CreatedAt` (timestamp): Creation time
- `UpdatedAt` (timestamp): Last modification time

**Validation Rules**:
- Name: 2-100 characters
- Code: Required
- Type: Must be one of (menu, button, api)
- Status: Must be 0 or 1

**Indexes**:
- Primary key on `id`
- Unique index on `code`
- Index on `parent_id` (for hierarchy queries)
- Index on `type` (for filtering by permission type)
- Composite index on `(status, sort)` (for active permission lists)

**State Transitions**:
- None (permissions are relatively static)

---

### APIKey

**Purpose**: API keys for programmatic access

**Table Name**: `api_keys`

**GORM Model Definition**:
```go
type APIKey struct {
    ID          string     `gorm:"column:id;primaryKey;type:varchar(36)"`
    Name        string     `gorm:"column:name;not null;type:varchar(100)"`
    Key         string     `gorm:"column:key;uniqueIndex;not null;type:varchar(255)"`
    Secret      string     `gorm:"column:secret;not null;type:varchar(255)"` // bcrypt hash
    UserID      string     `gorm:"column:user_id;index;type:varchar(36)"`
    Description string     `gorm:"column:description;type:text"`
    ExpiresAt   *time.Time `gorm:"column:expires_at;type:timestamp"` // Nullable
    Status      int        `gorm:"column:status;default:1;type:integer"`
    LastUsedAt  *time.Time `gorm:"column:last_used_at;type:timestamp"` // Nullable
    CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime"`

    // Relationships
    User *User `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

func (APIKey) TableName() string {
    return "api_keys"
}
```

**Fields**:
- `ID` (string, PK): UUID primary key
- `Name` (string, required): Key name/description
- `Key` (string, unique, required): API key (format: ak_xxxxx)
- `Secret` (string, required): API secret hash (format before hash: sk_xxxxx)
- `UserID` (string, FK): Owner user ID
- `Description` (string): Additional description
- `ExpiresAt` (timestamp, nullable): Expiration time
- `Status` (int): Key status (1=active, 0=disabled)
- `LastUsedAt` (timestamp, nullable): Last usage time
- `CreatedAt` (timestamp): Creation time
- `UpdatedAt` (timestamp): Last modification time

**Validation Rules**:
- Name: 3-100 characters
- Key: Must have "ak_" prefix
- ExpiresAt: Must be in the future if set

**Indexes**:
- Primary key on `id`
- Unique index on `key`
- Index on `user_id` (for user's API keys)
- Composite index on `(expires_at, status)` (for cleanup queries)

**State Transitions**:
- Active (status=1) → Disabled (status=0): Manual deactivation
- Active → Expired: Automatic when expires_at < now()

---

## Junction Tables

### UserRole

**Purpose**: Many-to-many relationship between Users and Roles

**Table Name**: `user_roles`

**GORM Model Definition**:
```go
type UserRole struct {
    UserID string `gorm:"column:user_id;primaryKey;type:varchar(36)"`
    RoleID string `gorm:"column:role_id;primaryKey;type:varchar(36)"`
}

func (UserRole) TableName() string {
    return "user_roles"
}
```

**Fields**:
- `UserID` (string, PK, FK): Reference to users.id
- `RoleID` (string, PK, FK): Reference to roles.id

**Indexes**:
- Composite primary key on `(user_id, role_id)`
- Index on `role_id` (for reverse lookups)

**Foreign Keys**:
- `user_id` → `users.id` ON DELETE CASCADE
- `role_id` → `roles.id` ON DELETE RESTRICT

---

### RolePermission

**Purpose**: Many-to-many relationship between Roles and Permissions

**Table Name**: `role_permissions`

**GORM Model Definition**:
```go
type RolePermission struct {
    RoleID       string `gorm:"column:role_id;primaryKey;type:varchar(36)"`
    PermissionID string `gorm:"column:permission_id;primaryKey;type:varchar(36)"`
}

func (RolePermission) TableName() string {
    return "role_permissions"
}
```

**Fields**:
- `RoleID` (string, PK, FK): Reference to roles.id
- `PermissionID` (string, PK, FK): Reference to permissions.id

**Indexes**:
- Composite primary key on `(role_id, permission_id)`
- Index on `permission_id` (for reverse lookups)

**Foreign Keys**:
- `role_id` → `roles.id` ON DELETE CASCADE
- `permission_id` → `permissions.id` ON DELETE CASCADE

---

## GORM Configuration

### Global Configuration

```go
config := &gorm.Config{
    // Use custom logger with kart-io/logger
    Logger: NewGormLogger(kartLogger),

    // Naming strategy (use existing table/column names)
    NamingStrategy: schema.NamingStrategy{
        SingularTable: true, // Use singular table names
    },

    // Disable AutoMigrate in production
    DisableForeignKeyConstraintWhenMigrating: false, // Keep FK constraints
}
```

### Environment-Specific Settings

**Development/Test**:
```go
if cfg.Server.Mode != "release" {
    db.AutoMigrate(
        &User{}, &Role{}, &Permission{},
        &UserRole{}, &RolePermission{}, &APIKey{},
    )
}
```

**Production**:
```go
// AutoMigrate disabled
// Manual migration scripts used for schema changes
```

---

## Query Patterns

### Common GORM Queries

**User with Roles**:
```go
var user User
db.Preload("Roles").First(&user, "id = ?", userID)
```

**User Permissions (with joins)**:
```go
var permissions []Permission
db.Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
   Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
   Where("user_roles.user_id = ? AND permissions.status = ?", userID, 1).
   Find(&permissions)
```

**Paginated User List**:
```go
var users []User
var total int64

db.Model(&User{}).Where("status = ?", 1).Count(&total)
db.Where("status = ?", 1).
   Limit(pageSize).
   Offset((page - 1) * pageSize).
   Find(&users)
```

**Permission Tree (with children)**:
```go
var permissions []Permission
db.Preload("Children", "status = ?", 1).
   Where("parent_id IS NULL AND status = ?", 1).
   Order("sort ASC").
   Find(&permissions)
```

---

## Migration Notes

### Existing Data Compatibility

All GORM models are designed to work with existing data:
- Pointer types for nullable columns (ExpiresAt, LastUsedAt)
- Exact column name matching via `column:` tags
- Preserved foreign key constraints
- No schema modifications required

### Validation Before Deployment

Pre-deployment script will:
1. Generate GORM schema definition
2. Compare with production database schema
3. Report any mismatches
4. Require manual approval if differences found

### Rollback Strategy

If GORM migration fails:
1. Restore previous service version
2. Verify data integrity (record counts)
3. No schema rollback needed (no changes made)
4. Investigate GORM query issues
5. Fix and redeploy
