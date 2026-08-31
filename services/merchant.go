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
	DB                       *gorm.DB
	MerchantRepo             *repositories.MerchantRepository
	BankService              *BankService
	ComplianceService        *ComplianceService
	MerchantUserService      *MerchantUserService
	LinkedBankAccountService *LinkedBankAccountService
	UserService              *UserService
	WalletService            *WalletService
	NotificationService      *notifications.EmailService
}

func NewMerchantService(
	db *gorm.DB,
	merchantRepo *repositories.MerchantRepository,
	bankSrvc *BankService,
	complianceSrvc *ComplianceService,
	merchantUserSrvc *MerchantUserService,
	linkedBankAccountSrvc *LinkedBankAccountService,
	userSrvc *UserService,
	walletSrvc *WalletService,
	notificationService *notifications.EmailService,
) *MerchantService {
	return &MerchantService{
		DB:                       db,
		MerchantRepo:             merchantRepo,
		BankService:              bankSrvc,
		ComplianceService:        complianceSrvc,
		MerchantUserService:      merchantUserSrvc,
		LinkedBankAccountService: linkedBankAccountSrvc,
		UserService:              userSrvc,
		WalletService:            walletSrvc,
		NotificationService:      notificationService,
	}
}

var validate = validator.New()

func (ms *MerchantService) GetMerchantByID(merchantID uuid.UUID) (*models.Merchant, error) {
	merchant, err := ms.MerchantRepo.GetMerchantByID(merchantID)
	if err != nil {
		return nil, err
	}
	return merchant, nil
}

func (ms *MerchantService) CreateInitialMerchant(merchantRequest dto.CreateMerchantOnboardingRequest) (*dto.CreateMerchantOnboardingResponse, error) {
	initialMerchantRequest := models.Merchant{
		MerchantName:     merchantRequest.MerchantName,
		Email:            merchantRequest.Email,
		MerchantType:     merchantRequest.MerchantType,
		PhoneNumber:      merchantRequest.PhoneNumber,
		Status:           "pending",
		ComplianceStatus: "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	initialMerchant, err := ms.MerchantRepo.CreateMerchant(initialMerchantRequest)
	if err != nil {
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

func (ms *MerchantService) ProcessMerchantOnboarding(
	request dto.CreateMerchantOnboardingRequest,
	merchantID uuid.UUID,
) {
	var (
		complianceResponse *dto.ComplianceCheckResponse
		errC               error
	)

	if request.MerchantType == constants.MerchantTypeIndividual {
		complianceResponse, errC =
			ms.ComplianceService.PerformIndividualComplianceCheck(request, merchantID)
	} else {
		complianceResponse, errC =
			ms.ComplianceService.PerformCorporateComplianceCheck(request, merchantID)
	}

	if errC != nil {
		return
	}

	updatedMerchant, err := ms.UpdateMerchantCompliance(
		merchantID,
		complianceResponse,
	)
	if err != nil {
		return
	}

	if updatedMerchant.ComplianceStatus ==
		constants.ComplianceStatusRejected {
		return
	}

	provisionedUser, err := ms.ProvisionMerchant(request, merchantID)
	if err != nil {
		return
	}

	errN := ms.NotificationService.SendMerchantWelcomeEmail(provisionedUser.User, provisionedUser.TemporaryPassword)
	if errN != nil {
		return
	}
}

func (ms *MerchantService) ProvisionMerchant(
	merchantRequest dto.CreateMerchantOnboardingRequest,
	merchantID uuid.UUID,
) (*MerchantProvisioningResult, error) {

	tx := ms.DB.Begin()

	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 1. Create Wallet
	txWalletService := NewWalletService(
		ms.WalletService.WalletRepo.WithTx(tx),
	)
	_, err := txWalletService.CreateWalletForMerchant(merchantID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 2. Create Bank Account
	txBankService := NewBankService(
		ms.BankService.BankRepo.WithTx(tx),
		ms.BankService.MerchantRepo.WithTx(tx),
		ms.BankService.LinkedBankAccountRepo.WithTx(tx),
	)

	bankAccountRequest := &dto.CreateBankAccountRequest{
		AccountName: merchantRequest.MerchantName,
		MerchantID:  merchantID,
	}

	merchantBankAccount, err := txBankService.CreateBankAccount(
		bankAccountRequest,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 3. Create Linked Bank Account
	linkedBankAccountRequest := &dto.CreateLinkedBankAccountRequest{
		MerchantID:    merchantID,
		BankAccountID: merchantBankAccount.ID,
		Type:          merchantBankAccount.AccountType,
		Status:        merchantBankAccount.Status,
	}

	txLinkedBankAccountService := NewLinkedBankAccountService(
		ms.LinkedBankAccountService.LinkedBankAccountRepo.WithTx(tx),
	)

	_, err = txLinkedBankAccountService.CreateLinkedBankAccount(
		linkedBankAccountRequest,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
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

	txUserService := NewUserService(
		ms.UserService.UserRepo.WithTx(tx),
	)

	user, err := txUserService.CreateUser(
		userRequest,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 6. Create Merchant User
	merchantUserRequest := &dto.CreateMerchantUserRequest{
		MerchantID: merchantID,
		UserID:     user.User.ID,
		Role:       "owner",
	}

	txMerchantUserService := NewMerchantUserService(
		ms.MerchantUserService.MerchantUserRepo.WithTx(tx),
	)
	_, err = txMerchantUserService.CreateMerchantUser(
		merchantUserRequest,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	txMerchantRepo := ms.MerchantRepo.WithTx(tx)
	err = txMerchantRepo.UpdateMerchantStatus(
		merchantID,
		"active",
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 8. Commit everything
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	provisionMerchantResult := &MerchantProvisioningResult{
		User:              user.User,
		TemporaryPassword: user.TemporaryPassword,
	}

	return provisionMerchantResult, nil
}

func (ms *MerchantService) UpdateMerchantCompliance(merchantID uuid.UUID, complianceResponse *dto.ComplianceCheckResponse) (*models.Merchant, error) {
	updatedMerchant, err := ms.MerchantRepo.UpdateMerchantCompliance(merchantID, complianceResponse)
	if err != nil {
		return nil, err
	}
	return updatedMerchant, nil
}
