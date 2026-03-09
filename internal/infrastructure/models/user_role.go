package models

import "time"

// user_role.go — explicit join model with extra fields
type UserRole struct {
	UserID     uint      `gorm:"primaryKey"`
	RoleID     uint      `gorm:"primaryKey"`
	AssignedAt time.Time `gorm:"not null;default:NOW()"`
	AssignedBy uint      // user ID of admin who assigned the role

	User User `gorm:"foreignKey:UserID"`
	Role Role `gorm:"foreignKey:RoleID"`
}
