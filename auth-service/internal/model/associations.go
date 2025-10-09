package model

// UserRole represents the many-to-many relationship between Users and Roles
type UserRole struct {
	UserID string `gorm:"column:user_id;primaryKey;type:varchar(36)"`
	RoleID string `gorm:"column:role_id;primaryKey;type:varchar(36)"`
}

// TableName returns the table name for UserRole
func (UserRole) TableName() string {
	return "user_roles"
}

// RolePermission represents the many-to-many relationship between Roles and Permissions
type RolePermission struct {
	RoleID       string `gorm:"column:role_id;primaryKey;type:varchar(36)"`
	PermissionID string `gorm:"column:permission_id;primaryKey;type:varchar(36)"`
}

// TableName returns the table name for RolePermission
func (RolePermission) TableName() string {
	return "role_permissions"
}
