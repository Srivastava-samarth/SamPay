package services

import (
	"errors"
	"strings"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	"github.com/Srivastava-samarth/sampay/dto"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ComplianceService struct {
	UserRepo *repositories.UserRepository
}

func NewComplianceService(UserRepo *repositories.UserRepository) *ComplianceService {
	return &ComplianceService{
		UserRepo: UserRepo,
	}
}

func (cs *ComplianceService) PerformIndividualComplianceCheck(
	request *dto.IndividualComplianceCheckRequest,
	merchantID uuid.UUID,
) (*dto.ComplianceCheckResponse, error) {

	errV := cs.ValidateIndividualMerchantRequest(request)
	if errV != nil {
		return nil, errV
	}
	name, country, dob := request.FirstName+" "+request.LastName, request.Country, request.DateOfBirth
	// Perform compliance checks based on the provided information
	// For example, you can check if the name is valid, if the country is allowed, and if the date of birth meets certain criteria.
	// You can also integrate with external compliance services or databases for more comprehensive checks.

	// transform request to SIP model
	sip := constants.SIPPerson{
		Name:        name,
		Country:     country,
		DateOfBirth: dob,
	}
	specialInterestPersonList := constants.SIPList

	// match the request against the SIP list
	for _, person := range specialInterestPersonList {
		if person.Name == sip.Name && person.Country == sip.Country && person.DateOfBirth == sip.DateOfBirth {
			// If a match is found, return a compliance check response indicating the merchant is flagged.
			return &dto.ComplianceCheckResponse{
				MerchantID:       merchantID,
				ComplianceStatus: constants.ComplianceStatusRejected,
				ComplianceDate:   time.Now(),
				ComplianceReason: "special_interest_person",
				KYCDate:          time.Now(),
				KYC: dto.KYCData{
					Status: "rejected",
					Reason: "Special Interest Person (SIP)",
				},
				ComplianceDetails: dto.ComplianceDetails{
					IndividualComplianceCheckRequest: request,
				},
				Country: request.Country,
			}, nil
		}
	}

	// check for restricted countries
	for _, restrictedCountry := range constants.RestrictedCountries {
		if sip.Country == restrictedCountry {
			return &dto.ComplianceCheckResponse{
				MerchantID:       merchantID,
				ComplianceStatus: constants.ComplianceStatusRejected,
				ComplianceDate:   time.Now(),
				ComplianceReason: "sanctioned_country",
				KYCDate:          time.Now(),
				KYC: dto.KYCData{
					Status: "rejected",
					Reason: "Merchant flagged due to being from a restricted country.",
				},
				ComplianceDetails: dto.ComplianceDetails{
					IndividualComplianceCheckRequest: request,
				},
				Country: request.Country,
			}, nil
		}
	}

	blockedUser, err := cs.UserRepo.GetBlockedUserByEmail(request.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if blockedUser {
		return &dto.ComplianceCheckResponse{
			MerchantID:       merchantID,
			ComplianceStatus: constants.ComplianceStatusRejected,
			ComplianceDate:   time.Now(),
			ComplianceReason: "blocked_user",
			KYCDate:          time.Now(),
			KYC: dto.KYCData{
				Status: "rejected",
				Reason: "Merchant flagged due to being a blocked user.",
			},
			ComplianceDetails: dto.ComplianceDetails{
				IndividualComplianceCheckRequest: request,
			},
			Country: request.Country,
		}, nil
	}

	// no issues found, return approved status
	return &dto.ComplianceCheckResponse{
		MerchantID:       merchantID,
		ComplianceStatus: constants.ComplianceStatusApproved,
		ComplianceDate:   time.Now(),
		ComplianceReason: "compliance_check_passed",
		KYCDate:          time.Now(),
		KYC: dto.KYCData{
			Status: "approved",
			Reason: "Merchant passed all compliance checks.",
		},
		ComplianceDetails: dto.ComplianceDetails{
			IndividualComplianceCheckRequest: request,
		},
		Country: request.Country,
	}, nil
}

func (cs *ComplianceService) PerformCorporateComplianceCheck(
	request *dto.CompanyComplianceCheckRequest,
	merchantID uuid.UUID,
) (*dto.ComplianceCheckResponse, error) {

	errV := cs.ValidateCorporateMerchantRequest(request)
	if errV != nil {
		return nil, errV
	}
	company := request
	now := time.Now()

	// Check company against AML blocked entities
	for _, amlEntity := range constants.AMLBlockedEntities {
		if strings.EqualFold(amlEntity.Name, company.LegalName) &&
			strings.EqualFold(amlEntity.Country, company.IncorporationCountry) {

			return &dto.ComplianceCheckResponse{
				MerchantID:       merchantID,
				ComplianceStatus: constants.ComplianceStatusRejected,
				ComplianceDate:   now,
				ComplianceReason: "aml_match",
				KYCDate:          now,
				KYC: dto.KYCData{
					Status: "rejected",
					Reason: "Company matched an AML blocked entity.",
				},
				ComplianceDetails: dto.ComplianceDetails{
					CompanyComplianceCheckRequest: request,
				},
				Country: request.IncorporationCountry,
			}, nil
		}
	}

	// Check company incorporation country
	for _, restrictedCountry := range constants.RestrictedCountries {
		if strings.EqualFold(
			company.IncorporationCountry,
			restrictedCountry,
		) {

			return &dto.ComplianceCheckResponse{
				MerchantID:       merchantID,
				ComplianceStatus: constants.ComplianceStatusRejected,
				ComplianceDate:   now,
				ComplianceReason: "sanctioned_country",
				KYCDate:          now,
				KYC: dto.KYCData{
					Status: "rejected",
					Reason: "Company is incorporated in a restricted country.",
				},
				ComplianceDetails: dto.ComplianceDetails{
					CompanyComplianceCheckRequest: request,
				},
				Country: request.IncorporationCountry,
			}, nil
		}
	}

	// Validate company legal name
	if strings.TrimSpace(company.LegalName) == "" {
		return &dto.ComplianceCheckResponse{
			MerchantID:       merchantID,
			ComplianceStatus: constants.ComplianceStatusRejected,
			ComplianceDate:   now,
			ComplianceReason: "invalid_legal_name",
			KYCDate:          now,
			KYC: dto.KYCData{
				Status: "rejected",
				Reason: "Company legal name is required.",
			},
			ComplianceDetails: dto.ComplianceDetails{
				CompanyComplianceCheckRequest: request,
			},
			Country: request.IncorporationCountry,
		}, nil
	}

	// Validate registration number
	if strings.TrimSpace(company.RegistrationNumber) == "" {
		return &dto.ComplianceCheckResponse{
			MerchantID:       merchantID,
			ComplianceStatus: constants.ComplianceStatusRejected,
			ComplianceDate:   now,
			ComplianceReason: "invalid_registration_number",
			KYCDate:          now,
			KYC: dto.KYCData{
				Status: "rejected",
				Reason: "Company registration number is required.",
			},
			ComplianceDetails: dto.ComplianceDetails{
				CompanyComplianceCheckRequest: request,
			},
			Country: request.IncorporationCountry,
		}, nil
	}

	// Validate tax ID
	if strings.TrimSpace(company.TaxID) == "" {
		return &dto.ComplianceCheckResponse{
			MerchantID:       merchantID,
			ComplianceStatus: constants.ComplianceStatusRejected,
			ComplianceDate:   now,
			ComplianceReason: "invalid_tax_id",
			KYCDate:          now,
			KYC: dto.KYCData{
				Status: "rejected",
				Reason: "Company tax ID is required.",
			},
			ComplianceDetails: dto.ComplianceDetails{
				CompanyComplianceCheckRequest: request,
			},
			Country: request.IncorporationCountry,
		}, nil
	}

	// Check merchant email against blocked users
	blockedUser, err := cs.UserRepo.GetBlockedUserByEmail(
		request.Email,
	)

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if blockedUser {
		return &dto.ComplianceCheckResponse{
			MerchantID:       merchantID,
			ComplianceStatus: constants.ComplianceStatusRejected,
			ComplianceDate:   now,
			ComplianceReason: "blocked_user",
			KYCDate:          now,
			KYC: dto.KYCData{
				Status: "rejected",
				Reason: "Merchant email is associated with a blocked user.",
			},
			ComplianceDetails: dto.ComplianceDetails{
				CompanyComplianceCheckRequest: request,
			},
			Country: request.IncorporationCountry,
		}, nil
	}

	// All compliance checks passed
	return &dto.ComplianceCheckResponse{
		MerchantID:       merchantID,
		ComplianceStatus: constants.ComplianceStatusApproved,
		ComplianceDate:   now,
		ComplianceReason: "compliance_check_passed",
		KYCDate:          now,
		KYC: dto.KYCData{
			Status: "approved",
			Reason: "Corporate merchant passed all compliance checks.",
		},
		ComplianceDetails: dto.ComplianceDetails{
			CompanyComplianceCheckRequest: request,
		},
		Country: request.IncorporationCountry,
	}, nil
}

func (cs *ComplianceService) ValidateIndividualMerchantRequest(
	individualRequest *dto.IndividualComplianceCheckRequest,
) error {
	if individualRequest == nil {
		return errors.New("individual merchant request is required")
	}

	if individualRequest.FirstName == "" ||
		individualRequest.LastName == "" ||
		individualRequest.Country == "" ||
		individualRequest.DateOfBirth == "" ||
		individualRequest.Email == "" {
		return errors.New("first name, last name, country, date of birth, and email are required")
	}

	if !isValidCountry(individualRequest.Country) {
		return errors.New("invalid country")
	}

	if err := validateDateOfBirth(individualRequest.DateOfBirth); err != nil {
		return err
	}
	return nil
}

func (cs *ComplianceService) ValidateCorporateMerchantRequest(
	corporateRequest *dto.CompanyComplianceCheckRequest,
) error {
	if corporateRequest == nil {
		return errors.New("corporate merchant request is required")
	}

	if corporateRequest.LegalName == "" ||
		corporateRequest.RegistrationNumber == "" ||
		corporateRequest.TaxID == "" ||
		corporateRequest.Email == "" ||
		corporateRequest.IncorporationCountry == "" {
		return errors.New("legal name, registration number, tax ID, email, and incorporation country are required")
	}

	if !isValidCountry(corporateRequest.IncorporationCountry) {
		return errors.New("invalid country")
	}

	return nil
}

func validateDateOfBirth(dobString string) error {
	dob, err := time.Parse("2006-01-02", dobString)
	if err != nil {
		return errors.New("date of birth must be in YYYY-MM-DD format")
	}

	now := time.Now()

	if dob.After(now) {
		return errors.New("date of birth cannot be in the future")
	}

	age := now.Year() - dob.Year()

	if now.Month() < dob.Month() ||
		(now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}

	if age < 18 {
		return errors.New("merchant must be at least 18 years old")
	}

	return nil
}

func isValidCountry(country string) bool {
	country = strings.TrimSpace(country)

	_, exists := constants.ValidCountries[strings.ToLower(country)]

	return exists
}
