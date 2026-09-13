package activities

import (
	"context"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/services"
	"github.com/google/uuid"
)

func (a *Registry) CreateUser(
	ctx context.Context,
	request *dto.CreateUserRequest,
) (*services.CreatedUserResult, error) {
	return a.UserService.CreateUser(request)
}

func (a *Registry) CreateMerchantUser(
	ctx context.Context,
	request *dto.CreateMerchantUserRequest,
) (*dto.CreateMerchantUserResponse, error) {
	return a.MerchantUserService.CreateMerchantUser(
		request,
	)
}

func (a *Registry) GetMerchantById(
	merchantID uuid.UUID,
) (*models.Merchant, error) {
	return a.MerchantRepo.GetMerchantByID(merchantID)
}

func (a *Registry) SendUserWelcomeEmail(
	ctx context.Context,
	name string,
	email string,
	merchantName string,
	role string,
	temporaryPassword string,
) error {

	return a.NotificationService.SendUserOnboardingEmail(
		name,
		email,
		merchantName,
		role,
		temporaryPassword,
	)
}
