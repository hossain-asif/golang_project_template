package models

import (
	"time"

	"gorm.io/gorm"
)

type TwoFAMethod string

const (
	TwoFAMethodOTP    TwoFAMethod = "otp"    // email / SMS one-time code
	TwoFAMethodTOTP   TwoFAMethod = "totp"   // authenticator app (Google Auth, Authy)
	TwoFAMethodBackup TwoFAMethod = "backup" // single-use backup code
)

// UserTwoFAToken stores 2FA codes tied to a user (one-to-many).
// One user can have multiple tokens (e.g. OTP + backup codes).
type UserTwoFAToken struct {
	gorm.Model
	UserID    uint        `gorm:"not null;index"`
	Token     string      `gorm:"size:512;not null"` // hashed OTP / TOTP secret / backup code
	Method    TwoFAMethod `gorm:"size:20;not null;default:'otp'"`
	ExpiresAt time.Time   `gorm:"not null"`           // short TTL for OTP (e.g. 5 min), long for TOTP secret
	UsedAt    *time.Time  `gorm:"default:null"`       // non-null = already consumed → replay protection
	Attempts  int         `gorm:"not null;default:0"` // brute-force counter
	IPAddress string      `gorm:"size:45"`            // IPv4 / IPv6 of requester

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
