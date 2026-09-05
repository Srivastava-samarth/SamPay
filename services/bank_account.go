package services

import (
	"errors"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BankService struct {
	DB                    *gorm.DB
	BankRepo              *repositories.BankRepository
	MerchantRepo          *repositories.MerchantRepository
	LinkedBankAccountRepo *repositories.LinkedBankAccountRepository
}

func NewBankService(
	db *gorm.DB,
	BankRepo *repositories.BankRepository,
	MerchantRepo *repositories.MerchantRepository,
	LinkedBankAccountRepo *repositories.LinkedBankAccountRepository) *BankService {
	return &BankService{
		DB:                    db,
		MerchantRepo:          MerchantRepo,
		BankRepo:              BankRepo,
		LinkedBankAccountRepo: LinkedBankAccountRepo,
	}
}

func (bs *BankService) CreateBankAccount(bankAccountRequest *dto.CreateBankAccountRequest) (*dto.CreateBankAccountResponse, error) {
	_, err := bs.MerchantRepo.GetMerchantByID(bankAccountRequest.MerchantID)
	if err != nil {
		return nil, err
	}

	linkedBankAccounts, err := bs.LinkedBankAccountRepo.GetAllBankAccountLinkedByMerchantID(bankAccountRequest.MerchantID)
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

	bankAccountForMerchant, err := bs.BankRepo.CreateBankAccount(createBankAccountPayload)
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

func (bs *BankService) CreateBankAccountAndLink(bankAccountRequest *dto.CreateBankAccountRequest) (*models.BankAccount, *models.LinkedBankAccount, error) {
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

	txBankRepo := bs.BankRepo.WithTx(tx)
	txLinkedBankAccount := bs.LinkedBankAccountRepo.WithTx(tx)

	bankAccountRequestPayload := &models.BankAccount{
		AccountName: bankAccountRequest.AccountName,
		AccountType: constants.BankAccountTypeSecondary,
	}
	bankAccount, errBA := txBankRepo.CreateBankAccount(bankAccountRequestPayload)
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

	linkedBankAccount, errLBA := txLinkedBankAccount.CreateLinkedBankAccount(createLinkedBankAccountRequest)
	if errLBA != nil {
		tx.Rollback()
		return nil, nil, errLBA
	}

	if err := tx.Commit().Error; err != nil {
		return nil, nil, err
	}

	return bankAccount, linkedBankAccount, nil

}

func (bs *BankService) UpdateBankAccountAndlink(updateBankAccountRequest *dto.UpdateBankAccountRequest) (*models.BankAccount, *models.LinkedBankAccount, error){
	updatedBankAccount, errUBA := bs.BankRepo.UpdateBankAccount(updateBankAccountRequest)
	if errUBA != nil{
		return nil, nil, errUBA
	}

	updatedLinkedBanAccount, errULBA := bs.LinkedBankAccountRepo.UpdateLinkedBankAccount(updatedBankAccount.ID, &dto.UpdateLinkedBankAccountRequest{
		Type: updateBankAccountRequest.AccountType,
		Status: updateBankAccountRequest.Status,
	})
	if errULBA != nil{
		return nil, nil, errULBA
	}

	return updatedBankAccount, updatedLinkedBanAccount, nil
}

func (bs *BankService) GetBankAccountsByMerchantID(merchantID uuid.UUID) ([]*models.BankAccount, error){
	linkedBankAccounts, errLBA := bs.LinkedBankAccountRepo.GetAllBankAccountLinkedByMerchantID(merchantID)
	if errLBA != nil{
		return nil, errLBA
	}
	if len(linkedBankAccounts) == 0{
		return nil, errors.New("No bank account found")
	}

	var bankAccounts []*models.BankAccount
	for _, linkedBankAccount := range linkedBankAccounts{
		bankAccount, errBA := bs.BankRepo.GetBankAccountByID(linkedBankAccount.BankAccountID)
		if errBA != nil{
			return nil, errBA
		}

		bankAccounts = append(bankAccounts, bankAccount)
	}

	return bankAccounts, nil
}

func (bs *BankService) GetBankAccount(bankAccountID uuid.UUID) (*models.BankAccount, error){
	if bankAccountID == uuid.Nil{
		return nil, errors.New("bank_account_id is empty")
	}

	bankAccount, errBA := bs.BankRepo.GetBankAccountByID(bankAccountID)
	if errBA != nil{
		return nil, errBA
	}

	return bankAccount, nil
}