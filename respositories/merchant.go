package repositories

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"gorm.io/gorm"
)

func CreateMerchant(request models.Merchant, db *gorm.DB) (*models.Merchant, error) {
	merchant := &models.Merchant{
		ID:                utils.GenerateUUID(),
		MerchantReference: utils.GenerateMerchantReference(),
		MerchantName:      request.MerchantName,
		Email:             request.Email,
		Status:            request.Status,
		PhoneNumber:       request.PhoneNumber,
		MerchantType:      request.MerchantType,
		ComplianceStatus:  request.ComplianceStatus,
		CreatedAt:         request.CreatedAt,
		UpdatedAt:         request.UpdatedAt,
	}
	err := db.Create(merchant).Error
	if err != nil {
		return nil, err
	}
	return merchant, nil
}

func GetMerchantByID(merchantID string, db *gorm.DB) (*models.Merchant, error) {
	var merchant models.Merchant
	err := db.Where("id = ?", merchantID).First(&merchant).Error
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

func UpdateMerchant(merchant *models.Merchant, db *gorm.DB) (*models.Merchant, error) {
	err := db.Save(merchant).Error
	if err != nil {
		return nil, err
	}
	return merchant, nil
}
