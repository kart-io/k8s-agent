package model

import (
	"time"
)

// Permission represents a system permission in the database.
type Permission struct {
	ID          string    `gorm:"column:id;primaryKey;type:varchar(36)" db:"id"`
	ParentID    *string   `gorm:"column:parent_id;index;type:varchar(36)" db:"parent_id"` // Nullable, pointer for GORM
	Name        string    `gorm:"column:name;not null;type:varchar(100)" db:"name"`
	Code        string    `gorm:"column:code;uniqueIndex;not null;type:varchar(100)" db:"code"`
	Type        string    `gorm:"column:type;not null;type:varchar(20)" db:"type"` // menu, button, api
	Path        string    `gorm:"column:path;type:varchar(255)" db:"path"`
	Method      string    `gorm:"column:method;type:varchar(10)" db:"method"`        // For API type (GET, POST, etc.)
	Component   string    `gorm:"column:component;type:varchar(255)" db:"component"` // For menu type
	Icon        string    `gorm:"column:icon;type:varchar(50)" db:"icon"`
	Sort        int       `gorm:"column:sort;default:0;type:integer" db:"sort"`
	Status      int       `gorm:"column:status;default:1;type:integer" db:"status"` // 1=active, 0=disabled
	Description string    `gorm:"column:description;type:text" db:"description"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" db:"updated_at"`

	// GORM Relationships (self-referencing for hierarchy, many-to-many with Roles)
	Parent   *Permission  `gorm:"foreignKey:ParentID;references:ID"`
	Children []Permission `gorm:"foreignKey:ParentID;references:ID"`
	Roles    []Role       `gorm:"many2many:role_permissions;foreignKey:ID;joinForeignKey:PermissionID;References:ID;joinReferences:RoleID"`
}

// TableName returns the table name for Permission.
func (Permission) TableName() string {
	return "permissions"
}
