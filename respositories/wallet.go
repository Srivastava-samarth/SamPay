package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type WalletRepository struct{
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) *WalletRepository{
	return &WalletRepository{
		db: db,
	}
}

func (wr *WalletRepository) WithTx(tx *gorm.DB) *WalletRepository {
	return &WalletRepository{
		db: tx,
	}
}

func(wr *WalletRepository) CreateWalletForMerchant(merchantID uuid.UUID) (*models.Wallets, error) {
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
