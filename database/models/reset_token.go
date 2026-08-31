package database

import (
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash string     `gorm:"type:text;not null;uniqueIndex"`
	ExpiresAt time.Time  `gorm:"not null"`
	UsedAt    time.Time `gorm:"default:null"`
	CreatedAt time.Time  `gorm:"not null;autoCreateTime"`
}
