package repositories

import (
	"encoding/json"
	"errors"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (mp *Repository) CreateMerchant(request models.Merchant) (*models.Merchant, error) {
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
	err := mp.DB.Create(merchant).Error
	if err != nil {
		return nil, err
	}
	return merchant, nil
}

func (mp *Repository) GetMerchantByID(merchantID uuid.UUID) (*models.Merchant, error) {
	var merchant *models.Merchant
	err := mp.DB.Where(
		"id = ?",
		merchantID).
		First(&merchant).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return merchant, nil
}

func (mp *Repository) GetMerchants() ([]*models.Merchant, error) {
	var merchants []*models.Merchant

	if err := mp.DB.Find(&merchants).Error; err != nil {
		return nil, err
	}
	return merchants, nil
}

func (mp *Repository) UpdateMerchantCompliance(
	merchantID uuid.UUID,
	complianceResponse *dto.ComplianceCheckResponse,
) (*models.Merchant, error) {

	var merchant models.Merchant

	err := mp.DB.Where("id = ?", merchantID).First(&merchant).Error
	if err != nil {
		return nil, err
	}

	kycJSON, err := json.Marshal(complianceResponse.KYC)
	if err != nil {
		return nil, err
	}

	complianceDetails, errCD := json.Marshal(complianceResponse.ComplianceDetails)
	if errCD != nil {
		return nil, errCD
	}

	merchant.ComplianceDate = &complianceResponse.ComplianceDate
	merchant.ComplianceStatus = complianceResponse.ComplianceStatus
	merchant.ComplianceReason = complianceResponse.ComplianceReason
	merchant.KYC = datatypes.JSON(kycJSON)
	merchant.KYCDate = &complianceResponse.KYCDate
	merchant.UpdatedAt = time.Now()
	merchant.ComplianceDetails = datatypes.JSON(complianceDetails)
	merchant.Country = complianceResponse.Country

	if err := mp.DB.Save(&merchant).Error; err != nil {
		return nil, err
	}

	return &merchant, nil
}

func (mp *Repository) UpdateMerchantStatus(merchantID uuid.UUID, status string) (*models.Merchant, error) {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	err := mp.DB.
		Model(&models.Merchant{}).
		Where("id = ?", merchantID).
		Updates(updates).Error

	if err != nil {
		return nil, err
	}

	var merchant *models.Merchant
	errM := mp.DB.Where("id = ?", merchantID).First(&merchant).Error
	if errM != nil {
		return nil, errM
	}

	return merchant, nil
}

func (mp *Repository) UpdateMerchant(request *dto.UpdateMerchantRequest, merchantID uuid.UUID) (*models.Merchant, error) {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if request.MerchantName != "" {
		updates["merchant_name"] = request.MerchantName
	}

	if request.PhoneNumber != "" {
		updates["phone_number"] = request.PhoneNumber
	}

	err := mp.DB.
		Model(&models.Merchant{}).
		Where("id = ?", merchantID).
		Updates(updates).Error

	if err != nil {
		return nil, err
	}

	var merchant *models.Merchant
	if err := mp.DB.
		Where("id = ?", merchantID).
		First(&merchant).
		Error; err != nil {
		return nil, err
	}

	return merchant, nil
}

func (mp *Repository) GetMerchantByEmail(email string) (*models.Merchant, error) {
	var merchant *models.Merchant
	err := mp.DB.
		Where("email = ?", email).
		First(&merchant).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return merchant, nil
}
