package database

import (
	"time"

	"github.com/google/uuid"
)

type IdempotencyKey struct {
	ID             uuid.UUID `json:"id" gorm:"primaryKey"`
	MerchantID     uuid.UUID `json:"merchant_id" gorm:"not null"`
	IdempotencyKey string    `json:"idempotency_key" gorm:"not null"`
	ResponseBody   []byte    `json:"response_body" gorm:"not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`
}
