package repositories

import (
	"errors"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (br *Repository) CreateBankAccount(
	bankAccount *models.BankAccount,
) (*models.BankAccount, error) {

	newBankAccount := &models.BankAccount{
		ID:            utils.GenerateUUID(),
		AccountNumber: utils.GenerateBankAccountNumber(),
		AccountName:   bankAccount.AccountName,
		BankName:      "SamPay Bank",
		IFSCCode:      "SAMP5917AY",
		AccountType:   bankAccount.AccountType,
		Balance:       decimal.Zero,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := br.DB.Create(newBankAccount).Error; err != nil {
		return nil, err
	}

	return newBankAccount, nil
}

func (br *Repository) GetBankAccountByID(ID uuid.UUID) (*models.BankAccount, error) {
	var bankAccount models.BankAccount
	err := br.DB.Where("id = ?", ID).First(&bankAccount).Error
	if err != nil {
		return nil, err
	}
	return &bankAccount, nil
}

func (br *Repository) UpdateBankAccount(bankAccountID uuid.UUID, request *dto.UpdateBankAccountRequest) (*models.BankAccount, error) {
	var oldBankAccountEntry *models.BankAccount
	if err := br.DB.Where("id = ? AND status = ?", bankAccountID, "active").First(&oldBankAccountEntry).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if request.Balance.GreaterThan(decimal.Zero) &&
		oldBankAccountEntry.Balance != request.Balance {
		updates["balance"] = request.Balance
	}

	if request.AccountName != "" &&
		oldBankAccountEntry.AccountName != request.AccountName {
		updates["account_name"] = request.AccountName
	}

	if request.AccountType != "" &&
		oldBankAccountEntry.AccountType != request.AccountType {
		updates["account_type"] = request.AccountType
	}

	if request.Status != "" &&
		oldBankAccountEntry.Status != request.Status {
		updates["status"] = request.Status
	}

	if len(updates) == 1 {
		return nil, errors.New("no updates to be done")
	}

	err := br.DB.
		Model(&models.BankAccount{}).
		Where("id = ? AND status = ?", bankAccountID, "active").
		Updates(updates).Error

	if err != nil {
		return nil, err
	}

	var updatedBankAccount *models.BankAccount
	if err := br.DB.Where("id = ?", bankAccountID).First(&updatedBankAccount).Error; err != nil {
		return nil, err
	}

	return updatedBankAccount, nil
}
