package services

import (
	"errors"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/notifications"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/utils"
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
	LedgerService            *LedgerService
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
	ledgerSrvc *LedgerService,
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
		LedgerService:            ledgerSrvc,
	}
}

func (ms *MerchantService) GetMerchantByID(merchantID uuid.UUID) (*models.Merchant, error) {
	if merchantID == uuid.Nil {
		return nil, errors.New("merchant_id is required")
	}

	merchant, err := ms.MerchantRepo.GetMerchantByID(merchantID)
	if err != nil {
		return nil, err
	}
	if merchant == nil {
		return nil, errors.New("merchant not found")
	}
	return merchant, nil
}

func (ms *MerchantService) GetMerchants() ([]*models.Merchant, error) {
	merchants, errM := ms.MerchantRepo.GetMerchants()
	if errM != nil {
		return nil, errM
	}
	return merchants, nil
}

func (ms *MerchantService) CreateInitialMerchant(merchantRequest *dto.CreateMerchantOnboardingRequest) (*dto.CreateMerchantOnboardingResponse, error) {
	if merchantRequest == nil {
		return nil, errors.New("request can't be empty")
	}

	initialMerchantRequest := models.Merchant{
		MerchantName:     merchantRequest.MerchantName,
		Email:            merchantRequest.Email,
		MerchantType:     merchantRequest.MerchantType,
		PhoneNumber:      merchantRequest.PhoneNumber,
		Status:           constants.MerchantStatusPending,
		ComplianceStatus: constants.ComplianceStatusPending,
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

func (ms *MerchantService) ProvisionMerchant(
	merchantRequest *dto.CreateMerchantOnboardingRequest,
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
		ms.DB,
		ms.WalletService.WalletRepo.WithTx(tx),
		ms.BankService.BankRepo.WithTx(tx),
		ms.LinkedBankAccountService.LinkedBankAccountRepo.WithTx(tx),
		ms.LedgerService,
	)
	_, err := txWalletService.CreateWalletForMerchant(merchantID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 2. Create Bank Account
	txBankService := NewBankService(
		ms.DB,
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

	user, err := ms.UserService.CreateUser(
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
	_, errM := txMerchantRepo.UpdateMerchantStatus(
		merchantID,
		constants.MerchantStatusActive,
	)
	if errM != nil {
		tx.Rollback()
		return nil, errM
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

func (ms *MerchantService) ValidateMerchantOnboardingRequest(r *dto.CreateMerchantOnboardingRequest) error {
	switch r.MerchantType {
	case "individual":
		if r.Individual == nil {
			return errors.New("individual details are required")
		}

		if r.Company != nil {
			return errors.New("company details are not allowed for individual merchant")

		}

	case "company":
		if r.Company == nil {
			return errors.New("company details are required")
		}

		if r.Individual != nil {
			return errors.New("individual details are not allowed for company merchant")
		}
	}

	return nil
}

func (ms *MerchantService) UpdateMerchant(
	request *dto.UpdateMerchantRequest,
	merchantID uuid.UUID,
) (*models.Merchant, error) {
	if request == nil {
		return nil, errors.New("request can't be empty")
	}

	if merchantID == uuid.Nil {
		return nil, errors.New("merchant_id is required")
	}

	updatedMerchant, errUM := ms.MerchantRepo.UpdateMerchant(request, merchantID)
	if errUM != nil {
		return nil, errUM
	}

	return updatedMerchant, nil
}

func (ms *MerchantService) UpdateMerchantStatus(
	status *string,
	merchantID uuid.UUID,
) (*models.Merchant, error) {

	var newMerchant *models.Merchant
	if status == nil {
		return nil, errors.New("merchant status is required")
	}

	if !utils.IsValidMerchantStatus(*status) {
		return nil, errors.New("invalid merchant status")
	}

	if merchantID == uuid.Nil {
		return nil, errors.New("merchant_id is required")
	}

	err := ms.DB.Transaction(func(tx *gorm.DB) error {
		merchantRepo := ms.MerchantRepo.WithTx(tx)
		merchantUserRepo := ms.MerchantUserService.MerchantUserRepo.WithTx(tx)
		userRepo := ms.UserService.UserRepo.WithTx(tx)

		oldMerchant, err := merchantRepo.GetMerchantByID(merchantID)
		if err != nil {
			return err
		}

		if oldMerchant.Status == *status {
			newMerchant = oldMerchant
			return nil
		}

		merchant, err := merchantRepo.UpdateMerchantStatus(
			merchantID,
			*status,
		)
		if err != nil {
			return err
		}

		newMerchant = merchant

		merchantUsers, err :=
			merchantUserRepo.GetMerchantUsersByMerchantID(merchantID)

		if err != nil {
			return err
		}

		if len(merchantUsers) == 0 {
			return errors.New("no user linked")
		}

		for _, merchantUser := range merchantUsers {
			_, err := userRepo.UpdateUserStatus(
				*status,
				merchantUser.UserID,
			)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newMerchant, nil
}

func (ms *MerchantService) UpdateMerchantKycInfo(
	request *dto.UpdateMerchantKycRequest,
	merchantID uuid.UUID,
) (*models.Merchant, error) {

	if merchantID == uuid.Nil {
		return nil, errors.New("merchant_id is required")
	}

	if request == nil {
		return nil, errors.New("request can't be empty")
	}

	merchant, err := ms.MerchantRepo.GetMerchantByID(merchantID)
	if err != nil {
		return nil, err
	}

	if merchant.Status != constants.MerchantStatusActive {
		return nil, errors.New("can't update mechant kyc info")
	}

	var (
		complianceResult *dto.ComplianceCheckResponse
	)

	if merchant.MerchantType == constants.MerchantTypeIndividual {

		if request.IndividualUpdate == nil {
			return nil, errors.New("individual update is required")
		}

		complianceIndividualRequest :=
			&dto.IndividualComplianceCheckRequest{
				FirstName:   request.IndividualUpdate.FirstName,
				LastName:    request.IndividualUpdate.LastName,
				DateOfBirth: request.IndividualUpdate.DateOfBirth,
				Email:       merchant.Email,
				Country:     request.IndividualUpdate.Country,
			}

		complianceResult, err = ms.ComplianceService.
			PerformIndividualComplianceCheck(
				complianceIndividualRequest,
				merchantID,
			)

	} else {

		if request.CompanyUpdate == nil {
			return nil, errors.New("company update is required")
		}

		complianceCorporateRequest :=
			&dto.CompanyComplianceCheckRequest{
				LegalName:            request.CompanyUpdate.LegalName,
				Email:                merchant.Email,
				RegistrationNumber:   request.CompanyUpdate.RegistrationNumber,
				IncorporationCountry: request.CompanyUpdate.IncorporationCountry,
				TaxID:                request.CompanyUpdate.TaxID,
			}

		complianceResult, err = ms.ComplianceService.
			PerformCorporateComplianceCheck(
				complianceCorporateRequest,
				merchantID,
			)
	}

	if err != nil {
		return nil, err
	}

	var updatedMerchant *models.Merchant

	err = ms.DB.Transaction(func(tx *gorm.DB) error {

		merchantRepo := ms.MerchantRepo.WithTx(tx)
		merchantUserRepo := ms.MerchantUserService.
			MerchantUserRepo.WithTx(tx)
		userRepo := ms.UserService.
			UserRepo.WithTx(tx)

		updatedComplianceMerchant, err :=
			merchantRepo.UpdateMerchantCompliance(
				merchantID,
				complianceResult,
			)

		if err != nil {
			return err
		}

		var merchantStatus string

		if complianceResult.ComplianceStatus ==
			constants.ComplianceStatusApproved {

			merchantStatus = constants.MerchantStatusActive

		} else {

			merchantStatus = constants.MerchantStatusSuspended
		}

		if updatedComplianceMerchant.Status == merchant.Status {
			updatedMerchant = updatedComplianceMerchant
			return nil
		}

		updatedMerchant, err =
			merchantRepo.UpdateMerchantStatus(
				merchantID,
				merchantStatus,
			)

		if err != nil {
			return err
		}

		merchantUsers, err :=
			merchantUserRepo.GetMerchantUsersByMerchantID(
				merchantID,
			)

		if err != nil {
			return err
		}

		if len(merchantUsers) == 0 {
			return errors.New("no user linked to merchant")
		}

		for _, merchantUser := range merchantUsers {

			if _, err :=
				userRepo.UpdateUserStatus(
					merchantStatus,
					merchantUser.UserID,
				); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedMerchant, nil
}

func (ms *MerchantService) GetMerchantByEmail(email string) (*models.Merchant, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	merchant, errM := ms.MerchantRepo.GetMerchantByEmail(email)
	if errM != nil {
		return nil, errM
	}

	return merchant, nil
}
