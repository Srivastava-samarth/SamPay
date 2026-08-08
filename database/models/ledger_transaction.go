package database

import (
	"time"

	"github.com/google/uuid"
)

type LedgerTransaction struct {
	ID               uuid.UUID `json:"id" gorm:"primaryKey"`
	TransactionRef   string    `json:"transaction_ref" gorm:"unique;not null"`
	Type             string    `json:"type" gorm:"not null"`
	ReferenceID      uuid.UUID `json:"reference_id" gorm:"not null"`
	Status           string    `json:"status" gorm:"default:posted;not null"`
	SettlementStatus string    `json:"settlement_status" gorm:"default:pending;not null"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`
}