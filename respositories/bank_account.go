package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)


func CreateBankAccount(bankAccount *dto.CreateBankAccountRequest, db *gorm.DB) (*models.BankAccount, error) {
	parseBankAccountPayload := &models.BankAccount{
		ID: utils.GenerateUUID(),
		AccountNumber: bankAccount.AccountNumber,
		AccountName: bankAccount.AccountName,
		BankName: bankAccount.BankName,
		IFSCCode: "SAMP5917AY",
		AccountType: bankAccount.AccountType,
		Balance: decimal.NewFromFloat(0.0),
		Status: "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := db.Create(&parseBankAccountPayload).Error
	if err != nil {
		return nil, err
	}

	return parseBankAccountPayload, nil
}

func GetBankAccountByID(ID uuid.UUID, db *gorm.DB) (*models.BankAccount, error) {
	var bankAccount models.BankAccount
	err := db.Where("id = ?", ID).First(&bankAccount).Error
	if err != nil {
		return nil, err
	}
	return &bankAccount, nil
}

func GetAllBankAccountsByMerchantID(merchantID uuid.UUID, db *gorm.DB) ([]*models.BankAccount, error){
	var bankAccounts []*models.BankAccount
	err := db.Where("merchant_id = ? and status = ?", merchantID, "active").Find(&bankAccounts).Error
	if err != nil {
		return nil, err
	}
	return bankAccounts, nil
} 