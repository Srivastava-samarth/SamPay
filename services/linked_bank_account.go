package services

import (
	"errors"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
)

func (lbs *Services) CreateLinkedBankAccount(
	linkedBankAccountRequest *dto.CreateLinkedBankAccountRequest,
	) (*dto.CreateLinkedBankAccountResponse, error) {
	if linkedBankAccountRequest == nil {
		return nil, errors.New("request is required")
	}

	if linkedBankAccountRequest.MerchantID == uuid.Nil {
		return nil, errors.New("merchant_id is required")
	}

	if linkedBankAccountRequest.BankAccountID == uuid.Nil {
		return nil, errors.New("bank_account_id is required")
	}

	if linkedBankAccountRequest.Type == "" {
		return nil, errors.New("type is required")
	}

	if linkedBankAccountRequest.Status == "" {
		return nil, errors.New("status is required")
	}

	createLinkedBankAccountRequestPayload := &models.LinkedBankAccount{
		MerchantID:    linkedBankAccountRequest.MerchantID,
		BankAccountID: linkedBankAccountRequest.BankAccountID,
		Type:          linkedBankAccountRequest.Type,
		Status:        linkedBankAccountRequest.Status,
	}

	linkedBankAccount, err := lbs.Repo.CreateLinkedBankAccount(createLinkedBankAccountRequestPayload)
	if err != nil {
		return nil, err
	}

	linkedBankAccountResponse := &dto.CreateLinkedBankAccountResponse{
		LinkedBankAccountID: linkedBankAccount.ID,
		BankAccountID:       linkedBankAccount.BankAccountID,
		MerchantID:          linkedBankAccount.MerchantID,
		Type:                linkedBankAccount.Type,
		Status:              linkedBankAccount.Status,
	}

	return linkedBankAccountResponse, nil
}
