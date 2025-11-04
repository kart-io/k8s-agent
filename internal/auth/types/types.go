package types

import "time"

// User represents a system user.
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

// Role represents a user role.
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

// Permission represents a system permission.
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

// UserRole represents the many-to-many relationship between users and roles.
type UserRole struct {
	UserID string `json:"user_id" gorm:"primaryKey"`
	RoleID string `json:"role_id" gorm:"primaryKey"`
}

// RolePermission represents the many-to-many relationship between roles and permissions.
type RolePermission struct {
	RoleID       string `json:"role_id" gorm:"primaryKey"`
	PermissionID string `json:"permission_id" gorm:"primaryKey"`
}

// APIKey represents an API key for programmatic access.
type APIKey struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"not null"`
	Key         string     `json:"key" gorm:"uniqueIndex;not null"`
	Secret      string     `json:"-" gorm:"not null"` // Secret hash, not returned
	UserID      string     `json:"user_id" gorm:"index"`
	Description string     `json:"description"`
	ExpiresAt   *time.Time `json:"expires_at"`
	Status      int        `json:"status" gorm:"default:1"` // 1: active, 0: disabled
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response.
type LoginResponse struct {
	Token     string    `json:"token"`
	JTI       string    `json:"jti"` // JWT ID for session tracking
	ExpiresAt time.Time `json:"expires_at"`
	User      *UserInfo `json:"user"`
}

// UserInfo represents user information (without sensitive data).
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RealName string `json:"real_name"`
	Avatar   string `json:"avatar"`
	Roles    []Role `json:"roles"`
}

// MenuTree represents a menu tree structure.
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

// PermissionCheck represents a permission check request.
type PermissionCheck struct {
	UserID string `json:"user_id"`
	Path   string `json:"path"`
	Method string `json:"method"`
}

// UserCreateRequest represents user creation request.
type UserCreateRequest struct {
	Username string   `json:"username" binding:"required,min=3,max=50"`
	Password string   `json:"password" binding:"required,min=8"`
	Email    string   `json:"email" binding:"required,email"`
	RealName string   `json:"real_name"`
	Phone    string   `json:"phone"`
	Avatar   string   `json:"avatar"`
	RoleIDs  []string `json:"role_ids"`
}

// UserUpdateRequest represents user update request.
type UserUpdateRequest struct {
	Email    string   `json:"email" binding:"omitempty,email"`
	RealName string   `json:"real_name"`
	Phone    string   `json:"phone"`
	Avatar   string   `json:"avatar"`
	Status   *int     `json:"status"` // Pointer to distinguish between zero value and not set
	RoleIDs  []string `json:"role_ids"`
}

// RoleRequest represents role create/update request.
type RoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Status      *int   `json:"status"`
	Sort        int    `json:"sort"`
}

// PermissionRequest represents permission create/update request.
type PermissionRequest struct {
	ParentID    string `json:"parent_id"`
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Code        string `json:"code" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=menu button api"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	Component   string `json:"component"`
	Icon        string `json:"icon"`
	Sort        int    `json:"sort"`
	Status      *int   `json:"status"`
	Description string `json:"description"`
}

// APIKeyCreateRequest represents API key creation request.
type APIKeyCreateRequest struct {
	Name        string    `json:"name" binding:"required,min=3,max=100"`
	Description string    `json:"description"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// APIKeyWithSecret represents API key with secret (returned only once).
type APIKeyWithSecret struct {
	APIKey
	SecretPlain string `json:"secret"` // Plain secret, shown only once
}

// PaginationParams represents pagination parameters.
type PaginationParams struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Sort     string `json:"sort" form:"sort"`
	Order    string `json:"order" form:"order"` // asc or desc
}

// PaginatedResponse represents paginated response.
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// MenuItem represents a menu item with hierarchy.
type MenuItem struct {
	ID        string      `json:"id"`
	ParentID  string      `json:"parent_id,omitempty"`
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	Component string      `json:"component"`
	Icon      string      `json:"icon"`
	Sort      int         `json:"sort"`
	Children  []*MenuItem `json:"children,omitempty"`
}

// PermissionNode represents a permission in tree structure.
type PermissionNode struct {
	ID       string            `json:"id"`
	ParentID string            `json:"parent_id,omitempty"`
	Name     string            `json:"name"`
	Code     string            `json:"code"`
	Type     string            `json:"type"`
	Path     string            `json:"path,omitempty"`
	Method   string            `json:"method,omitempty"`
	Icon     string            `json:"icon,omitempty"`
	Sort     int               `json:"sort"`
	Children []*PermissionNode `json:"children,omitempty"`
}
