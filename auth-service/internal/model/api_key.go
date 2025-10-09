package model

import (
	"time"
)

// APIKey represents an API key in the database
type APIKey struct {
	ID          string     `gorm:"column:id;primaryKey;type:varchar(36)" db:"id"`
	Name        string     `gorm:"column:name;not null;type:varchar(100)" db:"name"`
	Key         string     `gorm:"column:key;uniqueIndex;not null;type:varchar(255)" db:"key"` // Format: ak_xxxxx
	Secret      string     `gorm:"column:secret;not null;type:varchar(255)" db:"secret"`       // bcrypt hash of sk_xxxxx
	UserID      string     `gorm:"column:user_id;index;type:varchar(36)" db:"user_id"`
	Description string     `gorm:"column:description;type:text" db:"description"`
	ExpiresAt   *time.Time `gorm:"column:expires_at;type:timestamp" db:"expires_at"`     // Nullable, pointer for GORM
	Status      int        `gorm:"column:status;default:1;type:integer" db:"status"`     // 1=active, 0=disabled
	LastUsedAt  *time.Time `gorm:"column:last_used_at;type:timestamp" db:"last_used_at"` // Nullable, pointer for GORM
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime" db:"updated_at"`

	// GORM Relationships (belongs to User)
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

// TableName returns the table name for APIKey
func (APIKey) TableName() string {
	return "api_keys"
}
