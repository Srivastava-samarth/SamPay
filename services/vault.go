package services

import (
	"errors"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type VaultService struct {
	db                    *gorm.DB
	VaultRepo             *repositories.VaultRepository
	LinkedBankAccountRepo *repositories.LinkedBankAccountRepository
	LedgerService         *LedgerService
	BankService           *BankService
}

func NewVaultService(
	db *gorm.DB,
	vaultRepo *repositories.VaultRepository,
	linkedBankAccountRepo *repositories.LinkedBankAccountRepository,
	ledgerService *LedgerService,
	bankService *BankService,
) *VaultService {
	return &VaultService{
		db:                    db,
		VaultRepo:             vaultRepo,
		LinkedBankAccountRepo: linkedBankAccountRepo,
		LedgerService:         ledgerService,
		BankService:           bankService,
	}
}

func (vs *VaultService) CreateVault(request *dto.CreateVaultRequest) (*models.Vault, error) {
	if request == nil {
		return nil, errors.New("request payload can't be empty")
	}

	if !utils.IsValidVaultStatus(request.Status) {
		return nil, errors.New("status is not a valid status")
	}

	if !utils.IsValidVaultType(request.Type) {
		return nil, errors.New("vault type is not a valid type")
	}

	if request.Balance.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("balance must be greater than 0")
	}

	existingVault, errEV := vs.VaultRepo.GetVaultByType(request.Type)
	if errEV != nil {
		return nil, errEV
	}

	if existingVault != nil {
		return nil, errors.New("vault already exist")
	}

	vault, errV := vs.VaultRepo.CreateVault(request)
	if errV != nil {
		return nil, errV
	}

	return vault, nil
}

func (vs *VaultService) GetVaults() ([]*models.Vault, error) {
	vaults, errV := vs.VaultRepo.GetVaults()
	if errV != nil {
		return nil, errV
	}

	return vaults, nil
}

func (vs *VaultService) GetVaultByID(vaultId uuid.UUID) (*models.Vault, error) {
	if vaultId == uuid.Nil {
		return nil, errors.New("vault_id is required")
	}

	vault, errV := vs.VaultRepo.GetVault(vaultId)
	if errV != nil {
		return nil, errV
	}
	return vault, nil
}

func (vs *VaultService) UpdateVaultBalance(merchantID uuid.UUID, balance decimal.Decimal, vaultType string) (*models.Vault, error) {
	if balance.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("balance must be greater than zero")
	}

	linkedBankAccount, errBA := vs.LinkedBankAccountRepo.GetPrimaryBankAccountLinkedByMerchantID(merchantID)
	if errBA != nil {
		return nil, errBA
	}

	if linkedBankAccount == nil {
		return nil, errors.New("primary bank account not found")
	}

	bankAccount, errBA := vs.BankService.GetBankAccount(linkedBankAccount.BankAccountID)
	if errBA != nil {
		return nil, errBA
	}

	if bankAccount == nil {
		return nil, errors.New("bank account not found")
	}

	if bankAccount.Balance.LessThan(balance) {
		return nil, errors.New("insufficient balance in bank account")
	}

	updatedBankAccount, _, errUB := vs.BankService.UpdateBankAccountAndlink(&dto.UpdateBankAccountRequest{
		ID:      bankAccount.ID,
		Balance: bankAccount.Balance.Sub(balance),
	})
	if errUB != nil {
		return nil, errUB
	}

	vault, errV := vs.VaultRepo.GetVaultByType(vaultType)
	if errV != nil {
		return nil, errV
	}

	if vault == nil {
		return nil, errors.New("vault not found")
	}

	updatedBalance := vault.Balance.Add(balance)
	updatedVault, errUV := vs.VaultRepo.UpdateVaultBalance(updatedBalance, vaultType)
	if errUV != nil {
		return nil, errUV
	}

	_, err := vs.LedgerService.PostTransaction(
		vs.db,
		&dto.PostLedgerTransactionRequest{
			ReferenceID:      "internal-vault-topup-" + uuid.New().String(),
			Type:             constants.LedgerAccountTypeBankAccount,
			Status:           constants.TransactionStatusCompleted,
			SettlementStatus: constants.LedgerSettlementSettled,
			Currency:         "INR",
		},
		[]*models.LedgerEntry{
			{
				AccountType: constants.LedgerAccountTypeBankAccount,
				AccountID:   updatedBankAccount.ID,
				EntryType:   constants.LedgerEntryTypeDebit,
				Amount:      balance,
				Currency:    "INR",
			},
			{
				AccountType: constants.LedgerAccountTypeVault,
				AccountID:   vault.ID,
				EntryType:   constants.LedgerEntryTypeCredit,
				Amount:      balance,
				Currency:    "INR",
			},
		},
	)
	if err != nil {
		return nil, errors.New("failed to post ledger transaction")
	}

	return updatedVault, errUV
}

func (vs *VaultService) UpdateVaultStatus(status string, vaultID uuid.UUID) (*models.Vault, error) {
	if !utils.IsValidVaultStatus(status) {
		return nil, errors.New("status is not a valid status")
	}

	if vaultID == uuid.Nil {
		return nil, errors.New("vault_id is required")
	}

	existingVault, errEV := vs.VaultRepo.GetVault(vaultID)
	if errEV != nil {
		return nil, errEV
	}

	if existingVault == nil {
		return nil, errors.New("vault doen not exist")
	}

	if existingVault.Status == status {
		return existingVault, nil
	}

	updatedVault, errUV := vs.VaultRepo.UpdateVaultStatus(status, vaultID)
	if errUV != nil {
		return nil, errUV
	}

	return updatedVault, nil
}
