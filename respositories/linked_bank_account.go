package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"gorm.io/gorm"
)

func CreateLinkedBankAccount(linkedBankAccount *dto.CreateLinkedBankAccountRequest, db *gorm.DB) (*models.LinkedBankAccount, error){
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