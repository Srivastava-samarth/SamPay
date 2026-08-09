package database

import (
	"time"

	"github.com/google/uuid"
)

type Merchant struct {
	ID                uuid.UUID  `json:"id" gorm:"primaryKey"`
	MerchantReference string     `json:"merchant_reference" gorm:"unique"`
	MerchantName      string     `json:"merchant_name"`
	Email             string     `json:"email"`
	PhoneNumber       string    `json:"phone_number"`
	Status            string     `json:"status"`
	MerchantType      string     `json:"merchant_type" gorm:"not null"`
	KYC               []byte     `json:"-" gorm:"type:jsonb"`
	KYCDate           *time.Time `json:"kyc_date"`
	ComplianceStatus  string     `json:"compliance_status" gorm:"default:pending;not null"`
	ComplianceDate    *time.Time `json:"compliance_date"`
	ComplianceReason  string     `json:"compliance_reason"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
