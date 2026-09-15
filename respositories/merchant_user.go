package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
)

func (mur *Repository) CreateMerchantUser(merchantUser *models.MerchantUser) (*models.MerchantUser, error) {
	newMerchantUser := &models.MerchantUser{
		ID:         utils.GenerateUUID(),
		MerchantID: merchantUser.MerchantID,
		UserID:     merchantUser.UserID,
		Role:       merchantUser.Role,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := mur.DB.Create(&newMerchantUser).Error
	if err != nil {
		return nil, err
	}

	return newMerchantUser, nil
}

func (mur *Repository) GetMerchantUserByUserID(userID uuid.UUID) (*models.MerchantUser, error) {
	var merchantUser *models.MerchantUser
	err := mur.DB.Where("user_id = ?", userID).First(&merchantUser).Error
	if err != nil {
		return nil, err
	}
	return merchantUser, nil
}

func (mur *Repository) GetMerchantUsersByMerchantID(merchantID uuid.UUID) ([]*models.MerchantUser, error) {
	var merchantUsers []*models.MerchantUser
	err := mur.DB.
		Where("merchant_id = ?", merchantID).
		Find(&merchantUsers).
		Error
	if err != nil {
		return nil, err
	}

	return merchantUsers, nil
}
