package database

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BankAccount struct {
	ID            uuid.UUID       `json:"id" gorm:"primaryKey"`
	AccountNumber string          `json:"-" gorm:"not null"`
	AccountName   string          `json:"account_name" gorm:"not null"`
	BankName      string          `json:"bank_name" gorm:"not null"`
	IFSCCode      string          `json:"ifsc_code" gorm:"not null"`
	AccountType   string          `json:"account_type" gorm:"not null"`
	Balance       decimal.Decimal `json:"balance" gorm:"not null"`
	Status        string          `json:"status" gorm:"default:active;not null"`
	CreatedAt     time.Time       `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt     time.Time       `json:"updated_at" gorm:"autoUpdateTime;not null"`
}
