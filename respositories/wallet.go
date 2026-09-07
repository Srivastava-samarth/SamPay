package repositories

import (
	"errors"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type WalletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{
		db: db,
	}
}

func (wr *WalletRepository) WithTx(tx *gorm.DB) *WalletRepository {
	return &WalletRepository{
		db: tx,
	}
}

func (wr *WalletRepository) CreateWalletForMerchant(merchantID uuid.UUID) (*models.Wallets, error) {
	wallet := &models.Wallets{
		ID:               utils.GenerateUUID(),
		MerchantID:       merchantID,
		AvailableBalance: decimal.Zero,
		ReservedBalance:  decimal.Zero,
		Status:           "active",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := wr.db.Create(wallet).Error; err != nil {
		return nil, err
	}

	return wallet, nil
}

func (wr *WalletRepository) GetWalletByMerchantId(merchantId uuid.UUID) (*models.Wallets, error) {
	var wallet *models.Wallets
	err := wr.db.Where("merchant_id = ? AND status = ?", merchantId, "active").First(&wallet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (wr *WalletRepository) UpdateWalletStatus(merchantId uuid.UUID, status string) (*models.Wallets, error) {
	err := wr.db.
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
	errW := wr.db.Where("merchant_id = ?", merchantId).First(&wallet).Error
	if errW != nil {
		return nil, errW
	}

	return wallet, nil
}

func (wr *WalletRepository) UpdateWalletBalance(merchantID uuid.UUID, walletRequest *dto.UpdateWallletBalanceRequest ) (*models.Wallets, error){
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if walletRequest.AvailableBalance.GreaterThan(decimal.Zero){
		updates["available_balance"] = walletRequest.AvailableBalance
	}

	if walletRequest.ReservedBalance.GreaterThan(decimal.Zero){
		updates["reserved_balance"] = walletRequest.ReservedBalance
	}

	err := wr.db.
		Model(&models.Wallets{}).
		Where("merchant_id = ? AND status = ?", merchantID, "active").
		Updates(updates).Error

	if err != nil {
		return nil, err
	}

	var updatedWallet *models.Wallets
	if err := wr.db.Where("merchant_id = ? AND status = ?", merchantID, "active").First(&updatedWallet).Error; err!= nil{
		return nil, err
	}

	return updatedWallet, nil

}