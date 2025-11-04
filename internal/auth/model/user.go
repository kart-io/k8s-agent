package model

import (
	"time"
)

// User represents a system user in the database.
type User struct {
	ID        string    `gorm:"column:id;primaryKey;type:varchar(36)" db:"id"`
	Username  string    `gorm:"column:username;uniqueIndex;not null;type:varchar(50)" db:"username"`
	Password  string    `gorm:"column:password;not null;type:varchar(255)" db:"password"` // bcrypt hash
	Email     string    `gorm:"column:email;uniqueIndex;type:varchar(100)" db:"email"`
	RealName  string    `gorm:"column:real_name;type:varchar(50)" db:"real_name"`
	Phone     string    `gorm:"column:phone;type:varchar(20)" db:"phone"`
	Avatar    string    `gorm:"column:avatar;type:varchar(255)" db:"avatar"`
	Status    int       `gorm:"column:status;default:1;type:integer" db:"status"` // 1=active, 0=disabled
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" db:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" db:"updated_at"`

	// GORM Relationships (many-to-many with Roles, one-to-many with APIKeys)
	Roles   []Role   `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:UserID;References:ID;joinReferences:RoleID"`
	APIKeys []APIKey `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

// TableName returns the table name for User.
func (User) TableName() string {
	return "users"
}
