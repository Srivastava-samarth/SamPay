package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateMerchantRequest struct {
	MerchantName string `json:"merchant_name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	MerchantType string `json:"merchant_type" validate:"required"`
	PhoneNumber  string `json:"phone_number" validate:"required"`
}

type CreateMerchantInitialResponse struct {
	ID                uuid.UUID `json:"id"`
	MerchantReference string    `json:"merchant_reference"`
	MerchantName      string    `json:"merchant_name"`
	Email             string    `json:"email"`
	PhoneNumber       string    `json:"phone_number"`
	Status            string    `json:"status"`
	ComplianceStatus  string    `json:"compliance_status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type IndividualMerchantOnboardingRequest struct {
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	DateOfBirth string `json:"date_of_birth" validate:"required"`
	Country     string `json:"country" validate:"required"`
	TaxID       string `json:"tax_id" validate:"required"`
	Address     string `json:"address" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"required"`
}

type CompanyMerchantOnboardingRequest struct {
	LegalName            string `json:"legal_name" validate:"required"`
	RegistrationNumber   string `json:"registration_number" validate:"required"`
	IncorporationCountry string `json:"incorporation_country" validate:"required"`
	TaxID                string `json:"tax_id" validate:"required"`
	BusinessType         string `json:"business_type" validate:"required"`
	RegisteredAddress    string `json:"registered_address" validate:"required"`
	OwnerName            string `json:"owner_name" validate:"required"`
	OwnerTaxID           string `json:"owner_tax_id" validate:"required"`
}

type CreateMerchantResponse struct {
	ID                uuid.UUID `json:"id"`
	MerchantReference string    `json:"merchant_reference"`
	MerchantName      string    `json:"merchant_name"`
	Email             string    `json:"email"`
	PhoneNumber       string    `json:"phone_number"`
	Status            string    `json:"status"`
	OwnerUserID       uuid.UUID `json:"owner_user_id"`
	WalletID          uuid.UUID `json:"wallet_id"`
	KYC               KYCData   `json:"kyc" gorm:"type:jsonb"`
	KYCDate           time.Time `json:"kyc_date"`
	ComplianceStatus  string    `json:"compliance_status" gorm:"default:pending;not null"`
	ComplianceDate    time.Time `json:"compliance_date"`
	ComplianceReason  string    `json:"compliance_reason"`
	CreatedAt         time.Time `json:"created_at"`
}
