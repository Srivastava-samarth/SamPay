package database

import (
	"time"

	"github.com/google/uuid"
)

type LinkedBankAccount struct {
	ID            uuid.UUID `json:"id" gorm:"primaryKey"`
	MerchantID    uuid.UUID `json:"merchant_id" gorm:"not null"`
	BankAccountID uuid.UUID `json:"bank_account_id" gorm:"not null"`
	Type          string    `json:"type" gorm:"not null"`
	Status        string    `json:"status" gorm:"default:active;not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`
}
