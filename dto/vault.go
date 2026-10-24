package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

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

type VaultTopupRequest struct {
	ReferenecID uuid.UUID       `json:"reference_id" binding:"required"`
	VaultID    uuid.UUID       `json:"vault_id" binding:"required"`
	Amount     decimal.Decimal `json:"amount" binding:"required"`
	Currency   string          `json:"currency" binding:"required"`
}

type CreateVaultTopupEvent struct {
	EventID   uuid.UUID `json:"event_id" binding:"required"`
	EventType string    `json:"event_type" binding:"required"`
	Payload   *VaultTopupRequest `json:"payload" binding:"required"`
}