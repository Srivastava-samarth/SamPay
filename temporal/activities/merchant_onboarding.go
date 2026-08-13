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
		return a.ComplianceService.PerformIndividualComplianceCheck(
			request,
			merchantID,
		)
	}

	return a.ComplianceService.PerformCorporateComplianceCheck(
		request,
		merchantID,
	)
}

func (a *Registry) ProvisionMerchant(
	ctx context.Context,
	request dto.CreateMerchantOnboardingRequest,
	merchantID uuid.UUID,
) (*services.MerchantProvisioningResult, error) {

	return a.MerchantService.ProvisionMerchant(
		request,
		merchantID,
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
	return a.MerchantService.UpdateMerchantCompliance(
		merchantID, 
		complianceResponse,
	)
}
