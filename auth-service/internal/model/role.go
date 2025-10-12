package model

import (
	"time"
)

// Role represents a user role in the database
type Role struct {
	ID          string    `gorm:"column:id;primaryKey;type:varchar(36)" db:"id"`
	Name        string    `gorm:"column:name;uniqueIndex;not null;type:varchar(50)" db:"name"`
	Code        string    `gorm:"column:code;uniqueIndex;not null;type:varchar(50)" db:"code"`
	Description string    `gorm:"column:description;type:text" db:"description"`
	Status      int       `gorm:"column:status;default:1;type:integer" db:"status"` // 1=active, 0=disabled
	Sort        int       `gorm:"column:sort;default:0;type:integer" db:"sort"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" db:"updated_at"`

	// GORM Relationships (many-to-many with Users and Permissions)
	Users       []User       `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:RoleID;References:ID;joinReferences:UserID"`
	Permissions []Permission `gorm:"many2many:role_permissions;foreignKey:ID;joinForeignKey:RoleID;References:ID;joinReferences:PermissionID"`
}

// TableName returns the table name for Role
func (Role) TableName() string {
	return "roles"
}
