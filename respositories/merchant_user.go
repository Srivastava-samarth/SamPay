package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantUserRepository struct {
	db *gorm.DB
}

func NewMerchantUserRepository(db *gorm.DB) *MerchantUserRepository {
	return &MerchantUserRepository{
		db: db,
	}
}

func (wr *MerchantUserRepository) WithTx(tx *gorm.DB) *MerchantUserRepository {
	return &MerchantUserRepository{
		db: tx,
	}
}

func (mur *MerchantUserRepository) CreateMerchantUser(merchantUser *models.MerchantUser) (*models.MerchantUser, error) {
	newMerchantUser := &models.MerchantUser{
		ID:         utils.GenerateUUID(),
		MerchantID: merchantUser.MerchantID,
		UserID:     merchantUser.UserID,
		Role:       merchantUser.Role,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := mur.db.Create(&newMerchantUser).Error
	if err != nil {
		return nil, err
	}

	return newMerchantUser, nil
}

func (mur *MerchantUserRepository) GetMerchantUserByUserID(userID uuid.UUID) (*models.MerchantUser, error) {
	var merchantUser *models.MerchantUser
	err := mur.db.Where("user_id = ?", userID).First(&merchantUser).Error
	if err != nil {
		return nil, err
	}
	return merchantUser, nil
}

func (mur *MerchantUserRepository) GetMerchantUsersByMerchantID(merchantID uuid.UUID) ([]*models.MerchantUser, error) {
	var merchantUsers []*models.MerchantUser
	err := mur.db.
		Where("merchant_id = ?", merchantID).
		Find(&merchantUsers).
		Error
	if err != nil {
		return nil, err
	}

	return merchantUsers, nil
}
