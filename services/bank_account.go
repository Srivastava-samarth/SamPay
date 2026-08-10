package services

import (
	"errors"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"gorm.io/gorm"
)

func CreateBankAccount(bankAccountRequest *dto.CreateBankAccountRequest, db *gorm.DB) (*dto.CreateBankAccountResponse, error) {
	_, err := repositories.GetMerchantByID(bankAccountRequest.MerchantID, db)
	if err != nil {
		return nil, err
	}

	linkedBankAccounts, err := repositories.GetAllBankAccountLinkedByMerchantID(bankAccountRequest.MerchantID, db)
	if err != nil {
		return nil, err
	}

	if len(linkedBankAccounts) == 3 {
		return nil, errors.New("Already have 3 bank accounts with this merchnat id")
	}
	var accountType = constants.BankAccountTypePrimary
	for _, linkedBankAccount := range linkedBankAccounts {
		if linkedBankAccount.Type == constants.BankAccountTypePrimary {
			accountType = constants.BankAccountTypeSecondary
		}
	}

	createBankAccountPayload := &models.BankAccount{
		AccountName: bankAccountRequest.AccountName,
		AccountType: accountType,
	}

	bankAccountForMerchant, err := repositories.CreateBankAccount(createBankAccountPayload, db)
	if err != nil {
		return nil, err
	}

	bankAccountResponse := &dto.CreateBankAccountResponse{
		ID:            bankAccountForMerchant.ID,
		AccountNumber: bankAccountForMerchant.AccountNumber,
		AccountType:   bankAccountForMerchant.AccountType,
		AccountName:   bankAccountForMerchant.AccountName,
		BankName:      bankAccountForMerchant.BankName,
		MerchantID:    bankAccountRequest.MerchantID,
		Status:        bankAccountForMerchant.Status,
		IFSCCode:      bankAccountForMerchant.IFSCCode,
	}
	return bankAccountResponse, nil
}
