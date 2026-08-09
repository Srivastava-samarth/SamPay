package controllers

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	services "github.com/Srivastava-samarth/sampay/services"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var validate = validator.New()

func CreateInitialMerchant(request dto.CreateMerchantRequest, db *gorm.DB) (*dto.CreateMerchantInitialResponse, *error) {
	validationErr := validate.Struct(request)
	if validationErr != nil {
		return nil, &validationErr
	}

	merchantRequest := models.Merchant{
		MerchantName:     request.MerchantName,
		Email:            request.Email,
		Status:           "pending",
		ComplianceStatus: "pending",
		PhoneNumber:      request.PhoneNumber,
		MerchantType:     request.MerchantType,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	initialMerchant, err := repositories.CreateMerchant(merchantRequest, db)
	if err != nil {
		return nil, &err
	}

	response := &dto.CreateMerchantInitialResponse{
		ID:                initialMerchant.ID,
		MerchantReference: initialMerchant.MerchantReference,
		MerchantName:      initialMerchant.MerchantName,
		Email:             initialMerchant.Email,
		Status:            initialMerchant.Status,
		ComplianceStatus:  initialMerchant.ComplianceStatus,
		CreatedAt:         initialMerchant.CreatedAt,
		UpdatedAt:         initialMerchant.UpdatedAt,
		PhoneNumber:       initialMerchant.PhoneNumber,
	}

	return response, nil
}

func CreateIndividualMerchantOnboarding(request dto.IndividualMerchantOnboardingRequest, merchantID string, db *gorm.DB) (*dto.CreateMerchantResponse, *error) {
	validationErr := validate.Struct(request)
	if validationErr != nil {
		return nil, &validationErr
	}

	merchant, err := services.GetMerchantByID(merchantID, db)
	if err != nil {
		return nil, &err
	}

	// compliance logic for individual merchant onboarding can be added here
	complianceResponse, complianceErr := services.PerformIndividualComplianceCheck(&request, merchantID, merchant.Email, db)
	if complianceErr != nil {
		return nil, &complianceErr
	}

	wallet, walletErr := services.CreateWalletForMerchant(merchant.ID, db)
	if walletErr != nil {
		return nil, &walletErr
	}

	bankAccount, bankAccountErr := services.CreateBankAccount(merchant.ID, db)

	response := &dto.CreateMerchantResponse{
		ID:                merchant.ID,
		MerchantReference: merchant.MerchantReference,
		MerchantName:      merchant.MerchantName,
		Email:             merchant.Email,
		Status:            merchant.Status,
		PhoneNumber:       merchant.PhoneNumber,
		WalletID:          wallet.ID,
		KYC:               complianceResponse.KYC,
		KYCDate:           complianceResponse.KYCDate,
		ComplianceStatus:  complianceResponse.ComplianceStatus,
		ComplianceDate:    complianceResponse.ComplianceDate,
		ComplianceReason:  complianceResponse.ComplianceReason,
		CreatedAt:         merchant.CreatedAt,
	}

	return response, nil
}
