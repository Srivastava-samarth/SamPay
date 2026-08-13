package services

import (
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
)

type MerchantUserService struct{
	MerchantUserRepo *repositories.MerchantUserRepository
}

func NewMerchantUserService(
	MerchantUserRepo *repositories.MerchantUserRepository,
) *MerchantUserService{
	return &MerchantUserService{
		MerchantUserRepo: MerchantUserRepo,
	}
}

func(mus *MerchantUserService) CreateMerchantUser(merchantUserRequest *dto.CreateMerchantUserRequest) (*dto.CreateMerchantUserResponse, error){
	createMerchantUserPayload := &models.MerchantUser{
		MerchantID: merchantUserRequest.MerchantID,
		UserID: merchantUserRequest.UserID,
		Role: merchantUserRequest.Role,
	}

	merchantUser, errMU := mus.MerchantUserRepo.CreateMerchantUser(createMerchantUserPayload)
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