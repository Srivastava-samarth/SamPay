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
	request *dto.CreateMerchantOnboardingRequest,
	merchantID uuid.UUID,
) (*dto.ComplianceCheckResponse, error) {

	if request.MerchantType == constants.MerchantTypeIndividual {
		individualComplianceRequest := &dto.IndividualComplianceCheckRequest{
			FirstName:   request.Individual.FirstName,
			LastName:    request.Individual.LastName,
			Email:       request.Email,
			DateOfBirth: request.Individual.DateOfBirth,
			Country:     request.Individual.Country,
		}

		return a.Services.PerformIndividualComplianceCheck(
			individualComplianceRequest,
			merchantID,
		)
	}

	companyComplianceRequest := &dto.CompanyComplianceCheckRequest{
		LegalName:            request.Company.LegalName,
		Email:                request.Email,
		RegistrationNumber:   request.Company.RegistrationNumber,
		IncorporationCountry: request.Company.IncorporationCountry,
		TaxID:                request.Company.TaxID,
	}
	return a.Services.PerformCorporateComplianceCheck(
		companyComplianceRequest,
		merchantID,
	)
}

func (a *Registry) ProvisionMerchant(
	ctx context.Context,
	request *dto.CreateMerchantOnboardingRequest,
	merchantID uuid.UUID,
) (*services.MerchantProvisioningResult, error) {

	return a.Services.ProvisionMerchant(
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

func (a *Registry) SendKYCReattemptEmail(
	ctx context.Context,
	request *dto.CreateMerchantOnboardingRequest,
	merchantID uuid.UUID,
) error {

	return a.NotificationService.SendKYCReattemptEmail(
		request.Email,
		merchantID,
		request.MerchantType,
	)
}

func (a *Registry) UpdateMerchantCompliance(
	ctx context.Context,
	merchantID uuid.UUID,
	complianceResponse *dto.ComplianceCheckResponse,
) (*models.Merchant, error) {
	return a.Services.UpdateMerchantCompliance(
		merchantID,
		complianceResponse,
	)
}
