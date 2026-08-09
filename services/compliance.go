package services

import (
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/respositories"
	"gorm.io/gorm"
)

func PerformIndividualComplianceCheck(request *dto.IndividualMerchantOnboardingRequest, merchantID string, email string, db *gorm.DB) (*dto.ComplianceCheckResponse, error) {

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
			}, nil
		}
	}

	// check for restricted countries
	for _, restrictedCountry := range constants.RestrictedCountries {
		if sip.Country == restrictedCountry {
			return &dto.ComplianceCheckResponse{
				ComplianceStatus: constants.ComplianceStatusRejected,
				ComplianceDate:   time.Now(),
				ComplianceReason: "sanctioned_country",
				KYCDate:          time.Now(),
				KYC: dto.KYCData{
					Status: "rejected",
					Reason: "Merchant flagged due to being from a restricted country.",
				},
			}, nil
		}
	}

	blockedUser, err := repositories.GetBlockedUserByEmail(email, db)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if blockedUser {
		return &dto.ComplianceCheckResponse{
			ComplianceStatus: constants.ComplianceStatusRejected,
			ComplianceDate:   time.Now(),
			ComplianceReason: "blocked_user",
			KYCDate:          time.Now(),
			KYC: dto.KYCData{
				Status: "rejected",
				Reason: "Merchant flagged due to being a blocked user.",
			},
		}, nil
	}

	// no issues found, return approved status
	return &dto.ComplianceCheckResponse{
		ComplianceStatus: constants.ComplianceStatusApproved,
		ComplianceDate:   time.Now(),
		ComplianceReason: "compliance_check_passed",
		KYCDate:          time.Now(),
		KYC: dto.KYCData{
			Status: "approved",
			Reason: "Merchant passed all compliance checks.",
		},
	}, nil
}
