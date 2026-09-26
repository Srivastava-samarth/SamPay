package repositories

import (
	"errors"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (wr *Repository) CreateWalletForMerchant(
	merchantID uuid.UUID,
	) (*models.Wallets, error) {
	wallet := &models.Wallets{
		ID:               utils.GenerateUUID(),
		MerchantID:       merchantID,
		AvailableBalance: decimal.Zero,
		ReservedBalance:  decimal.Zero,
		Status:           "active",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := wr.DB.Create(wallet).Error; err != nil {
		return nil, err
	}

	return wallet, nil
}

func (wr *Repository) GetWalletByMerchantId(
	merchantId uuid.UUID,
	) (*models.Wallets, error) {
	var wallet *models.Wallets
	err := wr.DB.
	Clauses(clause.Locking{Strength: "UPDATE"}).
	Where("merchant_id = ? AND status = ?", merchantId, "active").First(&wallet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (wr *Repository) UpdateWalletStatus(
	merchantId uuid.UUID, 
	status string,
	) (*models.Wallets, error) {
	err := wr.DB.
		Model(&models.Wallets{}).
		Where("merchant_id = ?", merchantId).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
	if err != nil {
		return nil, err
	}

	var wallet *models.Wallets
	errW := wr.DB.Where("merchant_id = ?", merchantId).First(&wallet).Error
	if errW != nil {
		return nil, errW
	}

	return wallet, nil
}

func (wr *Repository) UpdateWalletBalance(
	merchantID uuid.UUID, 
	walletRequest *dto.UpdateWalletBalanceRequest,
	) (*models.Wallets, error) {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if walletRequest.AvailableBalance != nil {
		updates["available_balance"] = walletRequest.AvailableBalance
	}

	if walletRequest.ReservedBalance != nil {
		updates["reserved_balance"] = walletRequest.ReservedBalance
	}

	err := wr.DB.
		Model(&models.Wallets{}).
		Where("merchant_id = ? AND status = ?", merchantID, constants.MerchantStatusActive).
		Updates(updates).Error

	if err != nil {
		return nil, err
	}

	var updatedWallet *models.Wallets
	if err := wr.DB.Where("merchant_id = ? AND status = ?", merchantID, "active").First(&updatedWallet).Error; err != nil {
		return nil, err
	}

	return updatedWallet, nil

}
