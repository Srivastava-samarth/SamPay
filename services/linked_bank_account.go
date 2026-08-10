package services

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"gorm.io/gorm"
)

func CreateLinkedBankAccount(linkedBankAccountRequest *dto.CreateLinkedBankAccountRequest, db *gorm.DB ) (*dto.CreateLinkedBankAccountResponse, error){
	createLinkedBankAccountRequestPayload := &models.LinkedBankAccount{
		MerchantID: linkedBankAccountRequest.MerchantID,
		BankAccountID: linkedBankAccountRequest.BankAccountID,
		Type: linkedBankAccountRequest.Type,
		Status: linkedBankAccountRequest.Status,
	}

	linkedBankAccount, err := repositories.CreateLinkedBankAccount(createLinkedBankAccountRequestPayload, db)
	if err != nil{
		return nil, err
	}

	linkedBankAccountResponse := &dto.CreateLinkedBankAccountResponse{
		LinkedBankAccountID: linkedBankAccount.ID,
		BankAccountID: linkedBankAccount.BankAccountID,
		MerchantID: linkedBankAccount.MerchantID,
		Type: linkedBankAccount.Type,
		Status: linkedBankAccount.Status,
	}

	return linkedBankAccountResponse, nil
}