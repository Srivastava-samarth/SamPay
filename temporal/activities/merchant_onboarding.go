package activities

import (
	"context"

	"github.com/google/uuid"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/services"
)

func (a *Registry) PerformComplianceCheck(
	ctx context.Context,
	request dto.CreateMerchantOnboardingRequest,
	merchantID uuid.UUID,
) (*dto.ComplianceCheckResponse, error) {

	if request.MerchantType == constants.MerchantTypeIndividual {
		return services.PerformIndividualComplianceCheck(
			request,
			merchantID,
			a.DB,
		)
	}

	return services.PerformCorporateComplianceCheck(
		request,
		merchantID,
		a.DB,
	)
}

func (a *Registry) ProvisionMerchant(
	ctx context.Context,
	request dto.CreateMerchantOnboardingRequest,
	merchantID uuid.UUID,
) (*services.MerchantProvisioningResult, error) {

	return services.ProvisionMerchant(
		request,
		merchantID,
		a.DB,
	)
}

func (a *Registry) SendWelcomeEmail(
	ctx context.Context,
	user *dto.CreateUserResponse,
	temporaryPassword string,
) error {

	return a.NotificationService.SendMerchantWelcomeEmail(
		user,
		temporaryPassword,
	)
}

func (a *Registry) UpdateMerchantCompliance(
	ctx context.Context,
	merchantID uuid.UUID,
	complianceResponse *dto.ComplianceCheckResponse,
) (*models.Merchant, error) {
	return services.UpdateMerchantCompliance(
		merchantID, 
		complianceResponse, 
		a.DB,
	)
}
