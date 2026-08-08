package database

import (
	"github.com/google/uuid"
	"time"
)

type MerchantUser struct {
	ID         uuid.UUID `json:"id" gorm:"primaryKey"`
	MerchantID uuid.UUID `json:"merchant_id" gorm:"not null"`
	UserID     uuid.UUID `json:"user_id" gorm:"not null"`
	Role       string    `json:"role" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`
}
