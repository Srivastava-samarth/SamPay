package services

import (
	"context"
	"errors"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (bs *Services) UpdateBankAccountAndlinkStatus(
	ctx context.Context,
	bankAccountID uuid.UUID,
	updateBankAccountRequest *dto.UpdateBankAccountStatus,
	) (*models.BankAccount, *models.LinkedBankAccount, error) {
	if updateBankAccountRequest == nil {
		return nil, nil, errors.New("request is required")
	}

	if updateBankAccountRequest.Status == ""{
		return nil, nil, errors.New("status is required")
	}

	updatedBankAccount, errUBA := bs.Libs.UpdateBankAccountStatus(ctx, bankAccountID, updateBankAccountRequest)
	if errUBA != nil {
		return nil, nil, errUBA
	}

	updatedLinkedBanAccount, errULBA := bs.Repo.UpdateLinkedBankAccount(updatedBankAccount.ID, &dto.UpdateLinkedBankAccountRequest{
		Status: updateBankAccountRequest.Status,
	})
	if errULBA != nil {
		return nil, nil, errULBA
	}

	return updatedBankAccount, updatedLinkedBanAccount, nil
}

func (bs *Services) UpdateBankAccountBalance(
	ctx context.Context,
	bankAccountID uuid.UUID,
	request *dto.UpdateBankAccountBalance,
) (*models.BankAccount, error){
	if request == nil{
		return nil, errors.New("request can't be nil")
	}

	if request.Balance == decimal.Zero{
		return nil, errors.New("balance_is_required")
	}

	updatedBankAccount, errUBA := bs.Libs.UpdateBankAccountBalance(ctx, bankAccountID, request)
	if errUBA != nil{
		return nil, errUBA
	}

	return updatedBankAccount, nil
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

func (bs *Services) ValidateCreateBankAccountRequest(
	ctx context.Context,
	request *dto.CreateBankAccountRequest,
	merchantID uuid.UUID,
) error{
	if request == nil{
		return errors.New("request can't be empty")
	}

	if request.AccountName == ""{
		return errors.New("account_name is required")
	}

	if request.AccountType == ""{
		return errors.New("account_type is required")
	}

	linkedBankAccounts, errLBA := bs.Repo.GetAllBankAccountLinkedByMerchantID(merchantID)
	if errLBA != nil{
		return errLBA
	}

	if len(linkedBankAccounts) == 2{
		return errors.New("can't create more than two accounts")
	}

	if linkedBankAccounts[0].Type == request.AccountType{
		return errors.New("accoun_type already exists")
	}

	return nil
}