package models

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	Name        string `gorm:"size:100;unique;not null"` // "admin", "editor", "viewer"
	Description string `gorm:"size:255"`

	// Many-to-many back-reference (optional, for preloading)
	Users []UserRole `gorm:"foreignKey:RoleID"`
}

func (r Role) GetID() uint             { return r.ID }
func (r Role) GetCreatedAt() time.Time { return r.CreatedAt }
