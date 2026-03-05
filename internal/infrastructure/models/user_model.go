package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string `gorm:"size:255;not null"`
	Email    string `gorm:"size:255;unique;not null"`
	Password string `gorm:"size:255;not null"`
}

// Implement Cursorable interface — one-time, two lines per model
func (u User) GetID() uint             { return u.ID }
func (u User) GetCreatedAt() time.Time { return u.CreatedAt }
