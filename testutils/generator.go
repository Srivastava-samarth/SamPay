package testutils

import (
	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func GenerateTestMerchant() *models.Merchant {
	merchantName := "Test Merchant"
	email := uuid.NewString() + "@test.com"
	phoneNumber := "9999999999"
	status := constants.MerchantStatusActive
	merchantType := constants.MerchantTypeIndividual
	country := "IN"
	complianceStatus := "pending"

	return &models.Merchant{
		ID:                uuid.New(),
		MerchantReference: utils.GenerateMerchantReference(),
		MerchantName:      merchantName,
		Email:             email,
		PhoneNumber:       phoneNumber,
		Status:            status,
		MerchantType:      merchantType,
		Country:           country,
		ComplianceStatus:  complianceStatus,
	}
}

func GenerateTestWallet(merchantID uuid.UUID) *models.Wallets {
	status := constants.MerchantStatusActive

	return &models.Wallets{
		ID:               uuid.New(),
		MerchantID:       merchantID,
		AvailableBalance: decimal.NewFromInt(5000),
		ReservedBalance:  decimal.Zero,
		Status:           status,
	}
}

func GenerateTestBankAccount() *models.BankAccount {
	accountNumber := uuid.NewString()
	accountName := "Test Account"
	bankName := "Test Bank"
	ifscCode := "TEST0001234"
	accountType := "savings"
	status := constants.MerchantStatusActive

	return &models.BankAccount{
		ID:            uuid.New(),
		AccountNumber: accountNumber,
		AccountName:   accountName,
		BankName:      bankName,
		IFSCCode:      ifscCode,
		AccountType:   accountType,
		Balance:       decimal.NewFromInt(10000),
		Status:        status,
	}
}
