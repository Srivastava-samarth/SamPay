package database

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Wallets struct {
	ID               uuid.UUID       `json:"id" gorm:"primaryKey"`
	MerchantID       uuid.UUID       `json:"merchant_id" gorm:"unique;not null"`
	AvailableBalance decimal.Decimal `json:"available_balance" gorm:"not null"`
	ReservedBalance  decimal.Decimal `json:"reserved_balance" gorm:"not null"`
	Status           string          `json:"status" gorm:"default:active;not null"`
	CreatedAt        time.Time       `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt        time.Time       `json:"updated_at" gorm:"autoUpdateTime;not null"`
}
