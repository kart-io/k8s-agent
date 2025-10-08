package types

import "time"

// User represents a system user
type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"` // Password hash, not returned in JSON
	Email     string    `json:"email" gorm:"uniqueIndex"`
	RealName  string    `json:"real_name"`
	Phone     string    `json:"phone"`
	Avatar    string    `json:"avatar"`
	Status    int       `json:"status" gorm:"default:1"` // 1: active, 0: disabled
	RoleIDs   []string  `json:"role_ids" gorm:"-"`       // Not stored in user table
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Role represents a user role
type Role struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Code        string    `json:"code" gorm:"uniqueIndex;not null"`
	Description string    `json:"description"`
	Status      int       `json:"status" gorm:"default:1"` // 1: active, 0: disabled
	Sort        int       `json:"sort" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a system permission
type Permission struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	ParentID    string    `json:"parent_id" gorm:"index"`
	Name        string    `json:"name" gorm:"not null"`
	Code        string    `json:"code" gorm:"uniqueIndex;not null"`
	Type        string    `json:"type" gorm:"not null"` // menu, button, api
	Path        string    `json:"path"`                 // For menu/api
	Method      string    `json:"method"`               // For api: GET, POST, etc
	Component   string    `json:"component"`            // For menu: component path
	Icon        string    `json:"icon"`                 // For menu
	Sort        int       `json:"sort" gorm:"default:0"`
	Status      int       `json:"status" gorm:"default:1"` // 1: active, 0: disabled
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID string `json:"user_id" gorm:"primaryKey"`
	RoleID string `json:"role_id" gorm:"primaryKey"`
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	RoleID       string `json:"role_id" gorm:"primaryKey"`
	PermissionID string `json:"permission_id" gorm:"primaryKey"`
}

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Key         string    `json:"key" gorm:"uniqueIndex;not null"`
	Secret      string    `json:"-" gorm:"not null"` // Secret hash, not returned
	UserID      string    `json:"user_id" gorm:"index"`
	Description string    `json:"description"`
	ExpiresAt   time.Time `json:"expires_at"`
	Status      int       `json:"status" gorm:"default:1"` // 1: active, 0: disabled
	LastUsedAt  time.Time `json:"last_used_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expires_at"`
	User      *UserInfo   `json:"user"`
}

// UserInfo represents user information (without sensitive data)
type UserInfo struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	RealName string   `json:"real_name"`
	Avatar   string   `json:"avatar"`
	Roles    []Role   `json:"roles"`
}

// MenuTree represents a menu tree structure
type MenuTree struct {
	ID        string      `json:"id"`
	ParentID  string      `json:"parent_id"`
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	Component string      `json:"component"`
	Icon      string      `json:"icon"`
	Sort      int         `json:"sort"`
	Children  []*MenuTree `json:"children,omitempty"`
}

// PermissionCheck represents a permission check request
type PermissionCheck struct {
	UserID string `json:"user_id"`
	Path   string `json:"path"`
	Method string `json:"method"`
}
