package repositories

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)


func CreateBankAccount(bankAccount *dto.CreateBankAccountRequest, db *gorm.DB) (*models.BankAccount, error) {
	err := db.Create(&bankAccount).Error
	if err != nil {
		return nil, err
	}

	return &bankAccount, nil
}

func GetBankAccountByMerchantID(merchantID uuid.UUID, db *gorm.DB) (*models.BankAccount, error) {
	var bankAccount models.BankAccount
	err := db.Where("merchant_id = ?", merchantID).First(&bankAccount).Error
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