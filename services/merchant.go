package services

import (
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"gorm.io/gorm"
)

func GetMerchantByID(merchantID string, db *gorm.DB) (*models.Merchant, error) {
	merchant, err := repositories.GetMerchantByID(merchantID, db)
	if err != nil {
		return nil, err
	}
	return merchant, nil
}