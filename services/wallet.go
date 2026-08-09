package services

import (
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateWalletForMerchant(merchantID uuid.UUID, db *gorm.DB) (*models.Wallets, error) {
	wallet, walletErr := repositories.CreateWalletForMerchant(merchantID, db)
	if walletErr != nil {
		return nil, walletErr
	}
	return wallet, nil
}