package dto

import "github.com/shopspring/decimal"

type CreateVaultRequest struct {
	Type    string          `json:"type" gorm:"not null"`
	Balance decimal.Decimal `json:"balance" gorm:"not null"`
	Status  string          `json:"status" gorm:"default:active;not null"`
}
