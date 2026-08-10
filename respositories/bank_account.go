package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)


func CreateBankAccount(
	bankAccount *models.BankAccount,
	db *gorm.DB,
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

	if err := db.Create(newBankAccount).Error; err != nil {
		return nil, err
	}

	return newBankAccount, nil
}

func GetBankAccountByID(ID uuid.UUID, db *gorm.DB) (*models.BankAccount, error) {
	var bankAccount models.BankAccount
	err := db.Where("id = ?", ID).First(&bankAccount).Error
	if err != nil {
		return nil, err
	}
	return &bankAccount, nil
}