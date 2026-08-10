package repositories

import (
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"gorm.io/gorm"
)

func CreateMerchantUser(merchantUser *models.MerchantUser, db *gorm.DB) (*models.MerchantUser, error){
	newMerchantUser := &models.MerchantUser{
		ID: utils.GenerateUUID(),
		MerchantID: merchantUser.MerchantID,
		UserID: merchantUser.UserID,
		Role: merchantUser.Role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := db.Create(&newMerchantUser).Error
	if err != nil{
		return nil, err
	}

	return newMerchantUser, nil
}