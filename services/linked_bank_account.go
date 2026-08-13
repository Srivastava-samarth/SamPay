package services

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
)

type LinkedBankAccountService struct{
	LinkedBankAccountRepo *repositories.LinkedBankAccountRepository
}

func NewLinkedBankAccountService(
	LinkedBankAccountRepo *repositories.LinkedBankAccountRepository,
) *LinkedBankAccountService{
	return &LinkedBankAccountService{
		LinkedBankAccountRepo: LinkedBankAccountRepo,
	}
}

func(lbs *LinkedBankAccountService) CreateLinkedBankAccount(linkedBankAccountRequest *dto.CreateLinkedBankAccountRequest) (*dto.CreateLinkedBankAccountResponse, error){
	createLinkedBankAccountRequestPayload := &models.LinkedBankAccount{
		MerchantID: linkedBankAccountRequest.MerchantID,
		BankAccountID: linkedBankAccountRequest.BankAccountID,
		Type: linkedBankAccountRequest.Type,
		Status: linkedBankAccountRequest.Status,
	}

	linkedBankAccount, err := lbs.LinkedBankAccountRepo.CreateLinkedBankAccount(createLinkedBankAccountRequestPayload)
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