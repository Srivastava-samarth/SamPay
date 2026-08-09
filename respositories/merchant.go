package repositories

import (
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
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

func GetMerchantByID(merchantID uuid.UUID, db *gorm.DB) (*models.Merchant, error) {
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

func UpdateMerchantCompliance(
	merchantID uuid.UUID,
	complianceResponse *dto.ComplianceCheckResponse,
	db *gorm.DB,
) (*models.Merchant, error) {

	var merchant models.Merchant

	err := db.Where("id = ?", merchantID).First(&merchant).Error
	if err != nil {
		return nil, err
	}

	merchant.ComplianceDate = &complianceResponse.ComplianceDate
	merchant.ComplianceStatus = complianceResponse.ComplianceStatus
	merchant.ComplianceReason = complianceResponse.ComplianceReason
	merchant.KYC = complianceResponse.KYC
	merchant.KYCDate = &complianceResponse.KYCDate
	merchant.UpdatedAt = time.Now()

	if err := db.Save(&merchant).Error; err != nil {
		return nil, err
	}

	return &merchant, nil
}

func UpdateMerchantStatus(merchantID uuid.UUID, status string, db *gorm.DB) error{
	var merchant models.Merchant

	err := db.Where("id = ?", merchantID).First(&merchant).Error
	if err != nil {
		return err
	}

	merchant.Status = constants.MerchantStatusActive
	if err := db.Save(&merchant).Error; err != nil {
		return err
	}

	return nil
}