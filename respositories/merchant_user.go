package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"gorm.io/gorm"
)

type MerchantUserRepository struct{
	db *gorm.DB
}

func NewMerchantUserRepository(db *gorm.DB) *MerchantUserRepository{
	return &MerchantUserRepository{
		db: db,
	}
}

func (wr *MerchantUserRepository) WithTx(tx *gorm.DB) *MerchantUserRepository {
	return &MerchantUserRepository{
		db: tx,
	}
}

func(mur *MerchantUserRepository) CreateMerchantUser(merchantUser *models.MerchantUser) (*models.MerchantUser, error){
	newMerchantUser := &models.MerchantUser{
		ID: utils.GenerateUUID(),
		MerchantID: merchantUser.MerchantID,
		UserID: merchantUser.UserID,
		Role: merchantUser.Role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := mur.db.Create(&newMerchantUser).Error
	if err != nil{
		return nil, err
	}

	return newMerchantUser, nil
}