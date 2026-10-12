package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/security"
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
	email string,
	merchantID uuid.UUID,
	merchantType string,
) error {

	return a.NotificationService.SendKYCReattemptEmail(
		email,
		merchantID,
		merchantType,
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

func (a *Registry) CreateWallet(
	ctx context.Context,
	merchantID uuid.UUID,
) (*models.Wallets, error){
	return a.Services.CreateWalletForMerchant(merchantID)
}

func (a *Registry) ActivateMerchantOnboarding(
	ctx context.Context,
	merchantID uuid.UUID,
	userID uuid.UUID,
) (*models.User, error){
	_, errW := a.Repo.UpdateWalletStatus(merchantID, constants.MerchantStatusActive)
	if errW != nil{
		return nil, errW
	}

	user, errU := a.Repo.UpdateUserStatus(constants.UserStatusActive, userID)
	if errU != nil{
		return nil, errU
	}

	return user, nil
}

func (a *Registry) CreateLinkedBankAccount(
	ctx context.Context,
	bankAccount *models.BankAccount,
	merchantID uuid.UUID,
) (*models.LinkedBankAccount, error){
	return a.Repo.CreateLinkedBankAccount(&models.LinkedBankAccount{
		MerchantID: merchantID,
		BankAccountID: bankAccount.ID,
		Type: constants.BankAccountTypePrimary,
		Status: bankAccount.Status,
	})
}

func (a *Registry) PublishBankAccountCreationRequest(
	ctx context.Context,
	request dto.BankAccountCreationRequestEvent,
	workflowID string,
) (*models.Event,error){
	payload := request.Payload
	p, err := json.Marshal(payload)
	if err != nil {
		return nil,fmt.Errorf("marshal response payload: %w", err)
	}

	newEvent, errNE := a.CreateEvent(context.Background(), &dto.CreateEventRequest{
		EventID: request.EventID,
		EventType: request.EventType,
		Direction: string(constants.EventDirectionOutbound),
		Payload: p,
		WorkflowID: workflowID,
		Status: string(constants.EventStatusProcessing),
	})
	if errNE != nil{
		return nil, errNE
	}
	

	jsonPayload := datatypes.JSON(p)

	errP := a.nats.Publish(context.Background(), constants.EventType(newEvent.EventType), security.EventEnvelope{
		EventID: newEvent.EventID,
		EventType: newEvent.EventType,
		Payload: json.RawMessage(jsonPayload),
	})

	if errP != nil{
		return nil, errP
	}

	updatedEvent, errUE := a.UpdateEventStatus(context.Background(), newEvent.EventID, string(constants.EventStatusProcessed))
	if errUE != nil{
		return nil, errUE
	}

	return updatedEvent, nil
}

func (a *Registry) CreateEvent(
	ctx context.Context,
	eventPayload *dto.CreateEventRequest,
) (*models.Event, error){
	return a.Repo.CreateEvent(context.Background(), eventPayload)
}

func (a *Registry) UpdateEventStatus(
	ctx context.Context,
	eventID uuid.UUID,
	status string,
) (*models.Event, error){
	return a.Repo.UpdateEventStatus(context.Background(),eventID,status)
}

func (a *Registry) ProvisionUser(
	ctx context.Context,
	merchanTID uuid.UUID,
	request *dto.CreateMerchantOnboardingRequest,
) (*dto.MerchantProvisioningResult, error){
			var (
				firstName string
				lastName string
			)

			if request.MerchantType == constants.MerchantTypeIndividual{
				firstName = request.Individual.FirstName
				lastName = request.Individual.LastName
			}else{
				firstName = request.Company.OwnerFirstName
				lastName = request.Company.OwnerLastName
			}

			userRequest := &dto.CreateUserRequest{
				Email: request.Email,
				FirstName: firstName,
				LastName: lastName,
			}

			user, errU := a.Services.CreateUser(userRequest)
			if errU != nil{
				return nil, errU
			}

			merchantUserRequest := &dto.CreateMerchantUserRequest{
				MerchantID: merchanTID,
				UserID: user.User.ID,
				Role: "owner",
			}

			_, errMU := a.Services.CreateMerchantUser(merchantUserRequest)
			if errMU != nil{
				return nil, errMU
			}

			return &dto.MerchantProvisioningResult{
				User: user.User,
				TemporaryPassword: user.TemporaryPassword,
			}, nil
}
