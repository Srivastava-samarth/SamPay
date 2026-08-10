package services

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"gorm.io/gorm"
)

func CreateMerchantUser(merchantUserRequest *dto.CreateMerchantUserRequest, db *gorm.DB) (*dto.CreateMerchantUserResponse, error){
	createMerchantUserPayload := &models.MerchantUser{
		MerchantID: merchantUserRequest.MerchantID,
		UserID: merchantUserRequest.UserID,
		Role: merchantUserRequest.Role,
	}

	merchantUser, errMU := repositories.CreateMerchantUser(createMerchantUserPayload, db)
	if errMU != nil{
		return nil, errMU
	}

	merchantUserResponse := &dto.CreateMerchantUserResponse{
		MerchantUserLinkedID: merchantUser.ID,
		MerchantID: merchantUser.MerchantID,
		UserID: merchantUser.UserID,
		Role: merchantUser.Role,
	}

	return merchantUserResponse, nil
}