package repositories

import (
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
)

func (lbr *Repository) CreateLinkedBankAccount(
	linkedBankAccount *models.LinkedBankAccount,
	) (*models.LinkedBankAccount, error) {
	parseLinkedBankAccount := &models.LinkedBankAccount{
		ID:            utils.GenerateUUID(),
		MerchantID:    linkedBankAccount.MerchantID,
		BankAccountID: linkedBankAccount.BankAccountID,
		Type:          linkedBankAccount.Type,
		Status:        linkedBankAccount.Status,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	err := lbr.DB.Create(&parseLinkedBankAccount).Error
	if err != nil {
		return nil, err
	}

	return parseLinkedBankAccount, nil
}

func (lbr *Repository) GetAllBankAccountLinkedByMerchantID(
	merchantID uuid.UUID,
	) ([]*models.LinkedBankAccount, error) {
	var linkedBankAccounts []*models.LinkedBankAccount

	err := lbr.DB.Where(
		"merchant_id = ? AND status = ?",
		merchantID,
		"active",
	).Find(&linkedBankAccounts).Error

	if err != nil {
		return nil, err
	}

	return linkedBankAccounts, nil
}

func (br *Repository) UpdateLinkedBankAccount(
	bankAccountID uuid.UUID, 
	request *dto.UpdateLinkedBankAccountRequest,
	) (*models.LinkedBankAccount, error) {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if request.Type != "" {
		updates["type"] = request.Type
	}

	if request.Status != "" {
		updates["status"] = request.Status
	}

	err := br.DB.
		Model(&models.LinkedBankAccount{}).
		Where("bank_account_id = ?", bankAccountID).
		Updates(updates).Error
	if err != nil {
		return nil, err
	}

	var updatedLinkedBankAccount models.LinkedBankAccount
	err = br.DB.
		Where("bank_account_id = ?", bankAccountID).
		First(&updatedLinkedBankAccount).Error

	if err != nil {
		return nil, err
	}

	return &updatedLinkedBankAccount, nil
}

func (lbr *Repository) GetPrimaryBankAccountLinkedByMerchantID(
	merchantID uuid.UUID,
	) (*models.LinkedBankAccount, error) {
	var linkedBankAccount *models.LinkedBankAccount

	err := lbr.DB.Where(
		"merchant_id = ? AND status = ? AND type = ?",
		merchantID,
		"active",
		constants.BankAccountTypePrimary,
	).Find(&linkedBankAccount).Error

	if err != nil {
		return nil, err
	}
	return linkedBankAccount, nil

}

func (lbr *Repository) GetBankAccountLinkedByID(
	ID uuid.UUID,
	) (*models.LinkedBankAccount, error) {
	var linkedBankAccount *models.LinkedBankAccount

	err := lbr.DB.Where(
		"bank_account_id = ? AND status = ?",
		ID,
		"active",
	).Find(&linkedBankAccount).Error

	if err != nil {
		return nil, err
	}
	return linkedBankAccount, nil

}
