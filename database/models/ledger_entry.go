package database

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LedgerEntry struct {
	ID                  uuid.UUID       `json:"id" gorm:"primaryKey"`
	LedgerTransactionID uuid.UUID       `json:"ledger_transaction_id" gorm:"not null"`
	AccountType         string          `json:"account_type" gorm:"not null"`
	AccountID           uuid.UUID       `json:"account_id" gorm:"not null"`
	EntryType           string          `json:"entry_type" gorm:"not null"`
	Amount              decimal.Decimal `json:"amount" gorm:"not null"`
	Currency            string          `json:"currency" gorm:"not null"`
	CreatedAt           time.Time       `json:"created_at" gorm:"autoCreateTime;not null"`
}