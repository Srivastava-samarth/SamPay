package database

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Payout struct {
	ID                       uuid.UUID       `json:"id" gorm:"primaryKey"`
	MerchantID               uuid.UUID       `json:"merchant_id" gorm:"not null"`
	SourceWalletID           *uuid.UUID      `json:"source_wallet_id"`
	SourceBankAccountID      *uuid.UUID      `json:"source_bank_account_id"`
	DestinationBankAccountID uuid.UUID       `json:"destination_bank_account_id" gorm:"not null"`
	PayoutReference          string          `json:"payout_reference" gorm:"unique;not null"`
	ExternalReference        *string         `json:"external_reference"`
	Amount                   decimal.Decimal `json:"amount" gorm:"not null"`
	Currency                 string          `json:"currency" gorm:"not null"`
	Status                   string          `json:"status" gorm:"default:pending;not null"`
	Description              *string         `json:"description"`
	CreatedAt                time.Time       `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt                time.Time       `json:"updated_at" gorm:"autoUpdateTime;not null"`
}
