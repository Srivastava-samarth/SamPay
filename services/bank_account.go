package services

import (
	"errors"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
)

func (bs *Services) CreateBankAccount(
	bankAccountRequest *dto.CreateBankAccountRequest,
	) (*dto.CreateBankAccountResponse, error) {
	if bankAccountRequest == nil {
		return nil, errors.New("request is required")
	}

	if bankAccountRequest.AccountName == "" {
		return nil, errors.New("account_name is required")
	}

	if bankAccountRequest.MerchantID == uuid.Nil {
		return nil, errors.New("merchant_id is required")
	}

	_, err := bs.Repo.GetMerchantByID(bankAccountRequest.MerchantID)
	if err != nil {
		return nil, err
	}

	linkedBankAccounts, err := bs.Repo.GetAllBankAccountLinkedByMerchantID(bankAccountRequest.MerchantID)
	if err != nil {
		return nil, err
	}

	if len(linkedBankAccounts) == 2 {
		return nil, errors.New("already have 2 bank accounts with this merchnat id")
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

	bankAccountForMerchant, err := bs.Repo.CreateBankAccount(createBankAccountPayload)
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

func (bs *Services) CreateBankAccountAndLink(
	bankAccountRequest *dto.CreateBankAccountRequest,
	) (*models.BankAccount, *models.LinkedBankAccount, error) {
	tx := bs.DB.Begin()

	if tx.Error != nil {
		return nil, nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	repo := bs.Repo.WithTx(tx)

	linkedBankAccounts, err := repo.GetAllBankAccountLinkedByMerchantID(bankAccountRequest.MerchantID)
	if err != nil {
		return nil, nil, err
	}

	if len(linkedBankAccounts) == 2 {
		tx.Rollback()
		return nil, nil, errors.New("already have 2 bank accounts with this merchnat id")
	}

	bankAccountRequestPayload := &models.BankAccount{
		AccountName: bankAccountRequest.AccountName,
		AccountType: constants.BankAccountTypeSecondary,
	}
	bankAccount, errBA := repo.CreateBankAccount(bankAccountRequestPayload)
	if errBA != nil {
		tx.Rollback()
		return nil, nil, errBA
	}

	createLinkedBankAccountRequest := &models.LinkedBankAccount{
		MerchantID:    bankAccountRequest.MerchantID,
		BankAccountID: bankAccount.ID,
		Type:          bankAccount.AccountType,
		Status:        bankAccount.Status,
	}

	linkedBankAccount, errLBA := repo.CreateLinkedBankAccount(createLinkedBankAccountRequest)
	if errLBA != nil {
		tx.Rollback()
		return nil, nil, errLBA
	}

	if err := tx.Commit().Error; err != nil {
		return nil, nil, err
	}

	return bankAccount, linkedBankAccount, nil

}

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
		bankAccount, errBA := bs.Repo.GetBankAccountByID(linkedBankAccount.BankAccountID)
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

	bankAccount, errBA := bs.Repo.GetBankAccountByID(bankAccountID)
	if errBA != nil {
		return nil, errBA
	}

	return bankAccount, nil
}
