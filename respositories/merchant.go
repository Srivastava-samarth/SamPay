package repositories

import (
	"encoding/json"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type MerchantRepository struct{
	db *gorm.DB
}

func NewMerchantRepository(db *gorm.DB) *MerchantRepository{
	return &MerchantRepository{
		db: db,
	}
}

func (wr *MerchantRepository) WithTx(tx *gorm.DB) *MerchantRepository {
	return &MerchantRepository{
		db: tx,
	}
}

func(mp *MerchantRepository) CreateMerchant(request models.Merchant) (*models.Merchant, error) {
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
	err := mp.db.Create(merchant).Error
	if err != nil {
		return nil, err
	}
	return merchant, nil
}

func(mp *MerchantRepository) GetMerchantByID(merchantID uuid.UUID) (*models.Merchant, error) {
	var merchant models.Merchant
	err := mp.db.Where("id = ?", merchantID).First(&merchant).Error
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

func(mp *MerchantRepository) UpdateMerchant(merchant *models.Merchant) (*models.Merchant, error) {
	err := mp.db.Save(merchant).Error
	if err != nil {
		return nil, err
	}
	return merchant, nil
}

func(mp *MerchantRepository) UpdateMerchantCompliance(
	merchantID uuid.UUID,
	complianceResponse *dto.ComplianceCheckResponse,
) (*models.Merchant, error) {

	var merchant models.Merchant

	err := mp.db.Where("id = ?", merchantID).First(&merchant).Error
	if err != nil {
		return nil, err
	}

	kycJSON, err := json.Marshal(complianceResponse.KYC)
	if err != nil {
		return nil, err
	}

	merchant.ComplianceDate = &complianceResponse.ComplianceDate
	merchant.ComplianceStatus = complianceResponse.ComplianceStatus
	merchant.ComplianceReason = complianceResponse.ComplianceReason
	merchant.KYC = datatypes.JSON(kycJSON)
	merchant.KYCDate = &complianceResponse.KYCDate
	merchant.UpdatedAt = time.Now()

	if err := mp.db.Save(&merchant).Error; err != nil {
		return nil, err
	}

	return &merchant, nil
}

func(mp *MerchantRepository) UpdateMerchantStatus(merchantID uuid.UUID, status string) error{
	var merchant models.Merchant

	err := mp.db.Where("id = ?", merchantID).First(&merchant).Error
	if err != nil {
		return err
	}

	merchant.Status = status
	if err := mp.db.Save(&merchant).Error; err != nil {
		return err
	}

	return nil
}