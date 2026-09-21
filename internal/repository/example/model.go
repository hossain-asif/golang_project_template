package example

import (
	"gorm.io/gorm"
)

type Example struct {
	gorm.Model
	// rfu string `gorm:"size:255;not null"`
}
