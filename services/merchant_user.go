package services

import (
	"errors"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
)

func (mus *Services) CreateMerchantUser(merchantUserRequest *dto.CreateMerchantUserRequest) (*dto.CreateMerchantUserResponse, error) {
	if merchantUserRequest == nil {
		return nil, errors.New("request is required")
	}

	if merchantUserRequest.MerchantID == uuid.Nil {
		return nil, errors.New("merchant_id is required")
	}

	if merchantUserRequest.UserID == uuid.Nil {
		return nil, errors.New("user_id is required")
	}

	if merchantUserRequest.Role == "" {
		return nil, errors.New("role is required")
	}

	createMerchantUserPayload := &models.MerchantUser{
		MerchantID: merchantUserRequest.MerchantID,
		UserID:     merchantUserRequest.UserID,
		Role:       merchantUserRequest.Role,
	}

	merchantUser, errMU := mus.Repo.CreateMerchantUser(createMerchantUserPayload)
	if errMU != nil {
		return nil, errMU
	}

	merchantUserResponse := &dto.CreateMerchantUserResponse{
		MerchantUserLinkedID: merchantUser.ID,
		MerchantID:           merchantUser.MerchantID,
		UserID:               merchantUser.UserID,
		Role:                 merchantUser.Role,
	}

	return merchantUserResponse, nil
}
