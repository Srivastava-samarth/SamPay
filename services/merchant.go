package services

import (
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var validate = validator.New()

func GetMerchantByID(merchantID uuid.UUID, db *gorm.DB) (*models.Merchant, error) {
	merchant, err := repositories.GetMerchantByID(merchantID, db)
	if err != nil {
		return nil, err
	}
	return merchant, nil
}

func CreateMerchant(merchantRequest dto.CreateMerchantOnboardingRequest, db *gorm.DB) (*dto.CreateMerchantOnboardingResponse, error){
	validationErr := validate.Struct(merchantRequest)
	if validationErr != nil {
		return nil, validationErr
	}

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

	go ProcessMerchantOnboarding(merchantRequest,initialMerchant.ID,db)

	return nil, nil;
}

func ProcessMerchantOnboarding(request dto.CreateMerchantOnboardingRequest,  merchantID uuid.UUID, db *gorm.DB){
	var (
		complianceResponse *dto.ComplianceCheckResponse
		errC               error
	)
	if request.MerchantType == constants.MerchantTypeIndividual{
		complianceResponse ,errC = PerformIndividualComplianceCheck(request,merchantID,db)
	}else {
		complianceResponse ,errC = PerformIndividualComplianceCheck(request,merchantID,db)
	}

	if errC != nil {
		return
	}

	updatedMerchant, err := repositories.UpdateMerchantCompliance(
		merchantID,
		complianceResponse,
		db,
	)
	if err != nil {
		return
	}

	if updatedMerchant.ComplianceStatus == constants.ComplianceStatusRejected{
		return
	}

	errPM := ProvisionMerchant(request, merchantID, db);
	if errPM != nil{
		return 
	}

	err = repositories.UpdateMerchantStatus(
		merchantID,
		"active",
		db,
	)
	if err != nil {
		return
	}
}

func ProvisionMerchant(merchantRequest dto.CreateMerchantOnboardingRequest, merchantID uuid.UUID, db *gorm.DB) error{
	_, errW := CreateWalletForMerchant(merchantID,db)
	if errW != nil{
		return errW
	}

	merchantBankAccount, errBA := CreateBankAccount(merchantID,db)
	if errBA != nil{
		return errBA
	}

	_, errLBA := CreateLinkedBankAccount(*merchantBankAccount, merchantID, db)
	if errLBA != nil{
		return errLBA
	}

	return nil
}
