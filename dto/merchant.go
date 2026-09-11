package dto

import (
	"time"

	"github.com/google/uuid"
)

type IndividualMerchantOnboardingRequest struct {
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	DateOfBirth string `json:"date_of_birth" validate:"required"`
	Country     string `json:"country" validate:"required"`
	TaxID       string `json:"tax_id" validate:"required"`
	Address     string `json:"address" validate:"required"`
}

type CompanyMerchantOnboardingRequest struct {
	LegalName            string `json:"legal_name" validate:"required"`
	RegistrationNumber   string `json:"registration_number" validate:"required"`
	IncorporationCountry string `json:"incorporation_country" validate:"required"`
	TaxID                string `json:"tax_id" validate:"required"`
	BusinessType         string `json:"business_type" validate:"required"`
	RegisteredAddress    string `json:"registered_address" validate:"required"`
	OwnerFirstName       string `json:"owner_first_name" validate:"required"`
	OwnerLastName        string `json:"owner_last_name" validate:"required"`
	OwnerTaxID           string `json:"owner_tax_id" validate:"required"`
}

type CreateMerchantOnboardingResponse struct {
	ID                  uuid.UUID  `json:"id"`
	MerchantReference   string     `json:"merchant_reference"`
	MerchantName        string     `json:"merchant_name"`
	Email               string     `json:"email"`
	PhoneNumber         string     `json:"phone_number"`
	MerchantType        string     `json:"merchant_type"`
	Status              string     `json:"status"`
	ComplianceStatus    string     `json:"compliance_status"`
	OwnerUserID         *uuid.UUID `json:"owner_user_id,omitempty"`
	WalletID            *uuid.UUID `json:"wallet_id,omitempty"`
	LinkedBankAccountID *uuid.UUID `json:"linked_bank_account_id,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type CreateMerchantOnboardingRequest struct {
	MerchantType string `json:"merchant_type" binding:"required,oneof=individual company"`
	MerchantName string `json:"merchant_name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	PhoneNumber  string `json:"phone_number" binding:"required"`

	Individual *IndividualMerchantOnboardingRequest `json:"individual,omitempty"`
	Company    *CompanyMerchantOnboardingRequest    `json:"company,omitempty"`
}

type UpdateIndividualMerchantRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth" validate:"required"`
	Country     string `json:"country" validate:"required"`
	TaxID       string `json:"tax_id" validate:"required"`
	Address     string `json:"address" validate:"required"`
}

type UpdateCompanyMerchantRequest struct {
	LegalName            string `json:"legal_name"`
	RegistrationNumber   string `json:"registration_number"`
	IncorporationCountry string `json:"incorporation_country"`
	TaxID                string `json:"tax_id" validate:"required"`
	BusinessType         string `json:"business_type"`
	RegisteredAddress    string `json:"registered_address"`
	OwnerFirstName       string `json:"owner_first_name"`
	OwnerLastName        string `json:"owner_last_name"`
	OwnerTaxID           string `json:"owner_tax_id"`
}

type UpdateMerchantRequest struct {
	MerchantName     string                           `json:"merchant_name,omitempty"`
	PhoneNumber      string                           `json:"phone_number,omitempty"`
	IndividualUpdate *UpdateIndividualMerchantRequest `json:"individual_update,omitempty"`
	CompanyUpdate    *UpdateCompanyMerchantRequest    `json:"company_update,omitempty"`
}

type UpdateKYCRequest struct {
	Individual *IndividualMerchantOnboardingRequest `json:"individual,omitempty"`
	Company    *CompanyMerchantOnboardingRequest    `json:"company,omitempty"`
}