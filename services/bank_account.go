package services

import (
	"errors"
	"math/rand"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)


func CreateBankAccount(merchantID uuid.UUID, db *gorm.DB) (*models.BankAccount, error) {
	merchant, err := repositories.GetMerchantByID(string(merchantID), db)
	if err != nil{
		return nil, err
	}

	bankAccounts, err := repositories.GetAllBankAccountsByMerchantID(merchantID, db)
	if err != nil {
		return nil, err
	}

	if len(bankAccounts) == 3{
		return nil, errors.New("Already have 3 bank accounts with this merchnat id")
	}
	var accountType = constants.BankAccountTypePrimary;
	for _, bankAccount := range bankAccounts{
		if bankAccount.AccountType == constants.BankAccountTypePrimary{
			accountType = constants.BankAccountTypeSecondary
		}
	}

	createBankAccountRequest := &dto.CreateBankAccountRequest{
		BankName: "Sampay",
		AccountType: accountType,
		AccountNumber: *utils.GenerateBankAccountNumber(),
		AccountName: merchant.MerchantName,
		MerchantID: merchantID,
	}

	bankAccountForMerchant, err := repositories.CreateBankAccount(createBankAccountRequest, db)
	if err != nil{
		return nil, err
	}
	return bankAccountForMerchant, nil
}

