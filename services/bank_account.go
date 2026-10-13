package services

import (
	"context"
	"errors"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
)

func (bs *Services) UpdateBankAccountAndlink(
	updateBankAccountRequest *dto.UpdateBankAccountRequest,
	) (*models.BankAccount, *models.LinkedBankAccount, error) {
	if updateBankAccountRequest == nil {
		return nil, nil, errors.New("request is required")
	}

	if updateBankAccountRequest.ID == uuid.Nil {
		return nil, nil, errors.New("id is required")
	}

	updatedBankAccount, errUBA := bs.Repo.UpdateBankAccount(updateBankAccountRequest.ID, updateBankAccountRequest)
	if errUBA != nil {
		return nil, nil, errUBA
	}

	updatedLinkedBanAccount, errULBA := bs.Repo.UpdateLinkedBankAccount(updatedBankAccount.ID, &dto.UpdateLinkedBankAccountRequest{
		Type:   updateBankAccountRequest.AccountType,
		Status: updateBankAccountRequest.Status,
	})
	if errULBA != nil {
		return nil, nil, errULBA
	}

	return updatedBankAccount, updatedLinkedBanAccount, nil
}

func (bs *Services) GetBankAccountsByMerchantID(
	merchantID uuid.UUID,
	) ([]*models.BankAccount, error) {
	if merchantID == uuid.Nil {
		return nil, errors.New("merchant_id is required")
	}
	linkedBankAccounts, errLBA := bs.Repo.GetAllBankAccountLinkedByMerchantID(merchantID)
	if errLBA != nil {
		return nil, errLBA
	}
	if len(linkedBankAccounts) == 0 {
		return nil, errors.New("no bank account found")
	}

	var bankAccounts []*models.BankAccount
	for _, linkedBankAccount := range linkedBankAccounts {
		bankAccount, errBA := bs.Libs.GetBankAccount(context.Background(),linkedBankAccount.BankAccountID)
		if errBA != nil {
			return nil, errBA
		}

		bankAccounts = append(bankAccounts, bankAccount)
	}

	return bankAccounts, nil
}

func (bs *Services) GetBankAccount(
	bankAccountID uuid.UUID,
	) (*models.BankAccount, error) {
	if bankAccountID == uuid.Nil {
		return nil, errors.New("bank_account_id is empty")
	}

	bankAccount, errBA := bs.Libs.GetBankAccount(context.Background(), bankAccountID)
	if errBA != nil {
		return nil, errBA
	}

	return bankAccount, nil
}
