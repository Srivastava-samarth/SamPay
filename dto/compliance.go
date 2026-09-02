package dto

import (
	"time"

	"github.com/google/uuid"
)

type ComplianceCheckResponse struct {
	MerchantID       uuid.UUID `json:"merchant_id"`
	KYC              KYCData   `json:"kyc" gorm:"type:jsonb"`
	KYCDate          time.Time `json:"kyc_date"`
	ComplianceStatus string    `json:"compliance_status" gorm:"default:pending;not null"`
	ComplianceDate   time.Time `json:"compliance_date"`
	ComplianceReason string    `json:"compliance_reason"`
}

type KYCData struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type IndividualComplianceCheckRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email" validate:"required,email"`
	DateOfBirth string `json:"date_of_birth" validate:"required"`
	Country     string `json:"country" validate:"required"`
}

type CompanyComplianceCheckRequest struct {
	LegalName            string `json:"legal_name"`
	Email                string  `json:"email" validate:"required,email"`
	RegistrationNumber   string `json:"registration_number"`
	IncorporationCountry string `json:"incorporation_country"`
	TaxID                string  `json:"tax_id" validate:"required"`
}
