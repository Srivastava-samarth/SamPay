package dto

import "github.com/shopspring/decimal"

type CreateVaultRequest struct {
	Type    string          `json:"type" binding:"required"`
	Balance decimal.Decimal `json:"balance" binding:"required"`
	Status  string          `json:"status" binding:"required"`
}

type UpdateVaultBalance struct {
	Type    string          `json:"type" binding:"required"`
	Balance decimal.Decimal `json:"balance" binding:"required"`
}

type UpdateVaultStatus struct {
	Status string `json:"status" binding:"required"`
}
