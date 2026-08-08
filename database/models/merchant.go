package database

import (
	"time"

	"github.com/google/uuid"
)

type Merchant struct {
	ID                uuid.UUID `json:"id" gorm:"primaryKey"`
	MerchantReference string    `json:"merchant_reference" gorm:"unique"`
	MerchantName      string    `json:"merchant_name"`
	Email             string    `json:"email"`
	PhoneNumber       *string   `json:"phone_number"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
