package services

import (
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/notifications"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantProvisioningResult struct {
	User              *dto.CreateUserResponse
	TemporaryPassword string
}

type MerchantService struct {
	DB              *gorm.DB
	WorkflowStarter MerchantOnboardingWorkflowStarter
} 

func NewMerchantService(
	db *gorm.DB,
	workflowStarter MerchantOnboardingWorkflowStarter,
) *MerchantService {
	return &MerchantService{
		DB:              db,
		WorkflowStarter: workflowStarter,
	}
}

var validate = validator.New()

func GetMerchantByID(merchantID uuid.UUID, db *gorm.DB) (*models.Merchant, error) {
	merchant, err := repositories.GetMerchantByID(merchantID, db)
	if err != nil {
		return nil, err
	}
	return merchant, nil
}

func CreateInitialMerchant(merchantRequest dto.CreateMerchantOnboardingRequest, db *gorm.DB) (*dto.CreateMerchantOnboardingResponse, error){
	initialMerchantRequest := models.Merchant{
		MerchantName: merchantRequest.MerchantName,
		Email: merchantRequest.Email,
		MerchantType: merchantRequest.MerchantType,
		PhoneNumber: merchantRequest.PhoneNumber,
		Status: "pending",
		ComplianceStatus: "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	initialMerchant, err := repositories.CreateMerchant(initialMerchantRequest, db)
	if err != nil{
		return nil, err
	}

	response := &dto.CreateMerchantOnboardingResponse{
		ID:                initialMerchant.ID,
		MerchantReference: initialMerchant.MerchantReference,
		MerchantName:      initialMerchant.MerchantName,
		Email:             initialMerchant.Email,
		MerchantType:      initialMerchant.MerchantType,
		Status:            initialMerchant.Status,
		ComplianceStatus:  initialMerchant.ComplianceStatus,
		CreatedAt:         initialMerchant.CreatedAt,
		UpdatedAt:         initialMerchant.UpdatedAt,
	}

	return response, nil
}

// func CreateMerchant(merchantRequest dto.CreateMerchantOnboardingRequest, db *gorm.DB, notificationService *notifications.EmailService) (*dto.CreateMerchantOnboardingResponse, error){

// 	initialMerchantRequest := models.Merchant{
// 		MerchantName: merchantRequest.MerchantName,
// 		Email: merchantRequest.Email,
// 		MerchantType: merchantRequest.MerchantType,
// 		PhoneNumber: merchantRequest.PhoneNumber,
// 		Status: "pending",
// 		ComplianceStatus: "pending",
// 		CreatedAt: time.Now(),
// 		UpdatedAt: time.Now(),
// 	}

// 	initialMerchant, err := repositories.CreateMerchant(initialMerchantRequest, db)
// 	if err != nil{
// 		return nil, err
// 	}

// 	go ProcessMerchantOnboarding(merchantRequest,initialMerchant.ID,db,notificationService)

// 	response := &dto.CreateMerchantOnboardingResponse{
// 		ID:                initialMerchant.ID,
// 		MerchantReference: initialMerchant.MerchantReference,
// 		MerchantName:      initialMerchant.MerchantName,
// 		Email:             initialMerchant.Email,
// 		MerchantType:      initialMerchant.MerchantType,
// 		Status:            initialMerchant.Status,
// 		ComplianceStatus:  initialMerchant.ComplianceStatus,
// 		CreatedAt:         initialMerchant.CreatedAt,
// 		UpdatedAt:         initialMerchant.UpdatedAt,
// 	}

// 	return response, nil
// }

func ProcessMerchantOnboarding(
	request dto.CreateMerchantOnboardingRequest,
	merchantID uuid.UUID,
	db *gorm.DB,
	notificationService *notifications.EmailService,
) {
	var (
		complianceResponse *dto.ComplianceCheckResponse
		errC               error
	)

	if request.MerchantType == constants.MerchantTypeIndividual {
		complianceResponse, errC =
			PerformIndividualComplianceCheck(request, merchantID, db)
	} else {
		complianceResponse, errC =
			PerformIndividualComplianceCheck(request, merchantID, db)
	}

	if errC != nil {
		return
	}

	updatedMerchant, err := UpdateMerchantCompliance(
		merchantID,
		complianceResponse,
		db,
	)
	if err != nil {
		return
	}

	if updatedMerchant.ComplianceStatus ==
		constants.ComplianceStatusRejected {
		return
	}

	provisionedUser, err := ProvisionMerchant(request, merchantID, db)
	if err != nil {
		return
	}

	errN := notificationService.SendMerchantWelcomeEmail(provisionedUser.User,provisionedUser.TemporaryPassword)
	if errN != nil{
		return
	}
}

func ProvisionMerchant(
	merchantRequest dto.CreateMerchantOnboardingRequest,
	merchantID uuid.UUID,
	db *gorm.DB,
) (*MerchantProvisioningResult, error) {

	tx := db.Begin()

	if tx.Error != nil {
		return nil,tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 1. Create Wallet
	_, err := CreateWalletForMerchant(merchantID, tx)
	if err != nil {
		tx.Rollback()
		return nil,err
	}

	// 2. Create Bank Account
	bankAccountRequest := &dto.CreateBankAccountRequest{
		AccountName: merchantRequest.MerchantName,
		MerchantID: merchantID,
	}

	merchantBankAccount, err := CreateBankAccount(
		bankAccountRequest,
		tx,
	)
	if err != nil {
		tx.Rollback()
		return nil,err
	}

	// 3. Create Linked Bank Account
	linkedBankAccountRequest := &dto.CreateLinkedBankAccountRequest{
		MerchantID:    merchantID,
		BankAccountID: merchantBankAccount.ID,
		Type:          merchantBankAccount.AccountType,
		Status:        merchantBankAccount.Status,
	}

	_, err = CreateLinkedBankAccount(
		linkedBankAccountRequest,
		tx,
	)
	if err != nil {
		tx.Rollback()
		return nil,err
	}

	// 4. Determine owner name
	var firstName string
	var lastName string

	if merchantRequest.MerchantType ==
		constants.MerchantTypeIndividual {

		firstName = merchantRequest.Individual.FirstName
		lastName = merchantRequest.Individual.LastName

	} else {

		firstName = merchantRequest.Company.OwnerFirstName
		lastName = merchantRequest.Company.OwnerLastName
	}

	// 5. Create User
	userRequest := &dto.CreateUserRequest{
		Email:     merchantRequest.Email,
		FirstName: firstName,
		LastName:  lastName,
	}

	user, err := CreateUser(
		userRequest,
		tx,
	)
	if err != nil {
		tx.Rollback()
		return nil,err
	}

	// 6. Create Merchant User
	merchantUserRequest := &dto.CreateMerchantUserRequest{
		MerchantID: merchantID,
		UserID:     user.User.ID,
		Role:       "owner",
	}

	_, err = CreateMerchantUser(
		merchantUserRequest,
		tx,
	)
	if err != nil {
		tx.Rollback()
		return nil,err
	}

	err = repositories.UpdateMerchantStatus(
		merchantID,
		"active",
		tx,
	)
	if err != nil {
		tx.Rollback()
		return nil,err
	}

	// 8. Commit everything
	if err := tx.Commit().Error; err != nil {
		return nil,err
	}

	provisionMerchantResult := &MerchantProvisioningResult{
		User: user.User,
		TemporaryPassword: user.TemporaryPassword,
	}

	return provisionMerchantResult, nil
}

func UpdateMerchantCompliance(merchantID uuid.UUID,complianceResponse *dto.ComplianceCheckResponse, db *gorm.DB) (*models.Merchant, error){
	updatedMerchant, err := repositories.UpdateMerchantCompliance(merchantID,complianceResponse,db)
	if err != nil {
		return nil, err 
	}
	return updatedMerchant, nil
}