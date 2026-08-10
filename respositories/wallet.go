package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func CreateWalletForMerchant(merchantID uuid.UUID, db *gorm.DB) (*models.Wallets, error) {
	wallet := &models.Wallets{
		ID:               utils.GenerateUUID(),
		MerchantID:       merchantID,
		AvailableBalance: decimal.Zero,
		ReservedBalance:  decimal.Zero,
		Status:           "active",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.Create(wallet).Error; err != nil {
		return nil, err
	}

	return wallet, nil
}
