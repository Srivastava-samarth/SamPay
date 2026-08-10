package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateLinkedBankAccount(linkedBankAccount *models.LinkedBankAccount, db *gorm.DB) (*models.LinkedBankAccount, error){
	parseLinkedBankAccount := &models.LinkedBankAccount{
		ID: utils.GenerateUUID(),
		MerchantID: linkedBankAccount.MerchantID,
		BankAccountID: linkedBankAccount.BankAccountID,
		Type: linkedBankAccount.Type,
		Status: linkedBankAccount.Status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := db.Create(&parseLinkedBankAccount).Error
	if err != nil {
		return nil, err
	}

	return parseLinkedBankAccount, nil
}

func GetAllBankAccountLinkedByMerchantID(merchantID uuid.UUID, db *gorm.DB) ([]*models.LinkedBankAccount, error){
	var linkedBankAccounts []*models.LinkedBankAccount

	err := db.Where(
		"merchant_id = ? AND status = ?",
		merchantID,
		"active",
	).Find(&linkedBankAccounts).Error

	if err != nil {
		return nil, err
	}

	return linkedBankAccounts, nil
}