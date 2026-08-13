package services

import (
	"errors"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
)

type BankService struct {
	BankRepo              *repositories.BankRepository
	MerchantRepo          *repositories.MerchantRepository
	LinkedBankAccountRepo *repositories.LinkedBankAccountRepository
}

func NewBankService(BankRepo *repositories.BankRepository,
	MerchantRepo *repositories.MerchantRepository,
	LinkedBankAccountRepo *repositories.LinkedBankAccountRepository) *BankService {
	return &BankService{
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
