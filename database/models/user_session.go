package database

import (
	"time"

	"github.com/google/uuid"
)

type UserSession struct {
	ID               uuid.UUID  `json:"id" gorm:"primaryKey"`
	UserID           uuid.UUID  `json:"user_id" gorm:"not null"`
	RefreshTokenHash string     `json:"-"`
	ExpiresAt        time.Time  `json:"expires_at" gorm:"not null"`
	RevokedAt        *time.Time `json:"revoked_at"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime;not null"`
}
