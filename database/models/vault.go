package database

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Vault struct {
	ID        uuid.UUID       `json:"id" gorm:"primaryKey"`
	Type      string          `json:"type" gorm:"not null"`
	Balance   decimal.Decimal `json:"balance" gorm:"not null"`
	Status    string          `json:"status" gorm:"default:active;not null"`
	CreatedAt time.Time       `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt time.Time       `json:"updated_at" gorm:"autoUpdateTime;not null"`
}
