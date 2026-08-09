package services

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateLinkedBankAccount(BankAccount models.BankAccount, merchantID uuid.UUID, db *gorm.DB ) (*models.LinkedBankAccount, error){
	createLinkedBankAccountRequestPayload := &dto.CreateLinkedBankAccountRequest{
		MerchantID: merchantID,
		BankAccountID: BankAccount.ID,
		Type: BankAccount.AccountType,
		Status: BankAccount.Status,
	}

	linkedBankAccount, err := repositories.CreateLinkedBankAccount(createLinkedBankAccountRequestPayload, db)
	if err != nil{
		return nil, err
	}

	return linkedBankAccount, nil
}