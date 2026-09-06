package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LedgerTransactionRow struct {
	LedgerTransactionID uuid.UUID       `gorm:"column:ledger_transaction_id"`
	TransactionRef      string          `gorm:"column:transaction_ref"`
	Type                string          `gorm:"column:type"`
	ReferenceID         uuid.UUID       `gorm:"column:reference_id"`
	Status              string          `gorm:"column:status"`
	SettlementStatus    string          `gorm:"column:settlement_status"`
	Amount              decimal.Decimal `gorm:"column:amount"`
	Currency            string          `gorm:"column:currency"`
	EntryType           string          `gorm:"column:entry_type"`
	CreatedAt           time.Time       `gorm:"column:created_at"`
}

type PostLedgerTransactionRequest struct {
	Type             string          `gorm:"column:type"`
	ReferenceID      string          `gorm:"column:reference_id"`
	Status           string          `gorm:"column:status"`
	SettlementStatus string          `gorm:"column:settlement_status"`
	Amount           decimal.Decimal `gorm:"column:amount"`
	Currency         string          `gorm:"column:currency"`
	EntryType        string          `gorm:"column:entry_type"`
}

type PostLedgerEntryRequest struct {
	AccountType string          `json:"account_type" gorm:"not null"`
	AccountID   uuid.UUID       `json:"account_id" gorm:"not null"`
	EntryType   string          `json:"entry_type" gorm:"not null"`
	Amount      decimal.Decimal `json:"amount" gorm:"not null"`
	Currency    string          `json:"currency" gorm:"not null"`
}

type TransactionPagination struct {
	Records  int        `json:"records"`
	Next     *uuid.UUID `json:"next"`
	Previous *uuid.UUID `json:"previous"`
}
