package activities

import (
	"context"
	"testing"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/google/uuid"
)

func TestPerformComplianceCheck(t *testing.T) {
	registry := NewRegistry(
		db,
		testServices.NotificationService,
		testServices,
		testRepo,
	)

	t.Run("individual merchant - approved", func(t *testing.T) {
		cleanTestDB(t)

		merchantID := uuid.New()

		request := &dto.CreateMerchantOnboardingRequest{
			MerchantType: constants.MerchantTypeIndividual,
			Email:        "individual@example.com",
			Individual: &dto.IndividualMerchantOnboardingRequest{
				FirstName:   "John",
				LastName:    "Doe",
				DateOfBirth: "1995-05-10",
				Country:     "India",
			},
		}

		response, err := registry.PerformComplianceCheck(
			context.Background(),
			request,
			merchantID,
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if response == nil {
			t.Fatal("expected compliance response, got nil")
		}

		if response.MerchantID != merchantID {
			t.Errorf(
				"expected merchant ID %s, got %s",
				merchantID,
				response.MerchantID,
			)
		}

		if response.ComplianceStatus != constants.ComplianceStatusApproved {
			t.Errorf(
				"expected compliance status %s, got %s",
				constants.ComplianceStatusApproved,
				response.ComplianceStatus,
			)
		}

		if response.ComplianceReason != "compliance_check_passed" {
			t.Errorf(
				"expected compliance reason compliance_check_passed, got %s",
				response.ComplianceReason,
			)
		}
	})

	t.Run("corporate merchant - approved", func(t *testing.T) {
		cleanTestDB(t)

		merchantID := uuid.New()

		request := &dto.CreateMerchantOnboardingRequest{
			MerchantType: constants.MerrchantTypeCorporate,
			Email:        "company@example.com",
			Company: &dto.CompanyMerchantOnboardingRequest{
				LegalName:            "SamPay Technologies",
				RegistrationNumber:   "REG-12345",
				IncorporationCountry: "India",
				TaxID:                "TAX-12345",
			},
		}

		response, err := registry.PerformComplianceCheck(
			context.Background(),
			request,
			merchantID,
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if response == nil {
			t.Fatal("expected compliance response, got nil")
		}

		if response.MerchantID != merchantID {
			t.Errorf(
				"expected merchant ID %s, got %s",
				merchantID,
				response.MerchantID,
			)
		}

		if response.ComplianceStatus != constants.ComplianceStatusApproved {
			t.Errorf(
				"expected approved status, got %s",
				response.ComplianceStatus,
			)
		}

		if response.ComplianceReason != "compliance_check_passed" {
			t.Errorf(
				"expected reason compliance_check_passed, got %s",
				response.ComplianceReason,
			)
		}
	})

	t.Run("corporate merchant - AML match", func(t *testing.T) {
		cleanTestDB(t)

		merchantID := uuid.New()

		amlEntity := constants.AMLBlockedEntities[0]

		request := &dto.CreateMerchantOnboardingRequest{
			MerchantType: constants.MerrchantTypeCorporate,
			Email:        "company@example.com",
			Company: &dto.CompanyMerchantOnboardingRequest{
				LegalName:            amlEntity.Name,
				RegistrationNumber:   "REG-12345",
				IncorporationCountry: amlEntity.Country,
				TaxID:                "TAX-12345",
			},
		}

		response, err := registry.PerformComplianceCheck(
			context.Background(),
			request,
			merchantID,
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if response == nil {
			t.Fatal("expected compliance response, got nil")
		}

		if response.ComplianceStatus != constants.ComplianceStatusRejected {
			t.Errorf(
				"expected rejected status, got %s",
				response.ComplianceStatus,
			)
		}

		if response.ComplianceReason != "aml_match" {
			t.Errorf(
				"expected reason aml_match, got %s",
				response.ComplianceReason,
			)
		}
	})

	t.Run("individual merchant - invalid request", func(t *testing.T) {
		cleanTestDB(t)

		merchantID := uuid.New()

		request := &dto.CreateMerchantOnboardingRequest{
			MerchantType: constants.MerchantTypeIndividual,
			Email:        "",
			Individual: &dto.IndividualMerchantOnboardingRequest{
				FirstName:   "John",
				LastName:    "Doe",
				DateOfBirth: "1995-05-10",
				Country:     "India",
			},
		}

		_, err := registry.PerformComplianceCheck(
			context.Background(),
			request,
			merchantID,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err == nil {
			t.Fatal("expected underlying error, got nil")
		}
	})

	t.Run("corporate merchant - invalid request", func(t *testing.T) {
		cleanTestDB(t)

		merchantID := uuid.New()

		request := &dto.CreateMerchantOnboardingRequest{
			MerchantType: constants.MerrchantTypeCorporate,
			Email:        "",
			Company: &dto.CompanyMerchantOnboardingRequest{
				LegalName:            "SamPay Technologies",
				RegistrationNumber:   "REG-12345",
				IncorporationCountry: "India",
				TaxID:                "TAX-12345",
			},
		}

		_, err := registry.PerformComplianceCheck(
			context.Background(),
			request,
			merchantID,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err == nil {
			t.Fatal("expected underlying error, got nil")
		}
	})
}

func TestProvisionMerchant(t *testing.T) {
	t.Run("success - individual merchant", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		request := &dto.CreateMerchantOnboardingRequest{
			MerchantType: constants.MerchantTypeIndividual,
			MerchantName: "SamPay Test Merchant",
			Email:        "merchant@example.com",
			Individual: &dto.IndividualMerchantOnboardingRequest{
				FirstName: "John",
				LastName:  "Doe",
			},
		}

		result, err := registry.ProvisionMerchant(
			context.Background(),
			request,
			merchant.ID,
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("expected provisioning result, got nil")
		}

		if result.User == nil {
			t.Fatal("expected user in provisioning result, got nil")
		}

		if result.User.Email != request.Email {
			t.Errorf(
				"expected user email %s, got %s",
				request.Email,
				result.User.Email,
			)
		}

		if result.User.FirstName != "John" {
			t.Errorf(
				"expected first name John, got %s",
				result.User.FirstName,
			)
		}

		if result.User.LastName != "Doe" {
			t.Errorf(
				"expected last name Doe, got %s",
				result.User.LastName,
			)
		}

		if result.TemporaryPassword == "" {
			t.Error("expected temporary password to be generated")
		}

		// Verify wallet was created.
		var wallet models.Wallets

		if err := db.
			Where("merchant_id = ?", merchant.ID).
			First(&wallet).Error; err != nil {
			t.Fatalf("expected wallet to be created: %v", err)
		}

		// Verify bank account was created.
		var bankAccount models.BankAccount

		if err := db.
			Where("id IN (?)",
				db.Model(&models.LinkedBankAccount{}).
					Select("bank_account_id").
					Where("merchant_id = ?", merchant.ID),
			).
			First(&bankAccount).Error; err != nil {
			t.Fatalf("expected bank account to be created: %v", err)
		}

		// Verify linked bank account was created.
		var linkedBankAccount models.LinkedBankAccount

		if err := db.
			Where("merchant_id = ?", merchant.ID).
			First(&linkedBankAccount).Error; err != nil {
			t.Fatalf(
				"expected linked bank account to be created: %v",
				err,
			)
		}

		if linkedBankAccount.BankAccountID != bankAccount.ID {
			t.Errorf(
				"expected linked bank account to reference bank account %s, got %s",
				bankAccount.ID,
				linkedBankAccount.BankAccountID,
			)
		}

		// Verify merchant user was created.
		var merchantUser models.MerchantUser

		if err := db.
			Where("merchant_id = ? AND user_id = ?",
				merchant.ID,
				result.User.ID,
			).
			First(&merchantUser).Error; err != nil {
			t.Fatalf(
				"expected merchant user to be created: %v",
				err,
			)
		}

		if merchantUser.Role == "" || merchantUser.Role != "owner" {
			t.Errorf("expected merchant user role owner")
		}

		// Verify merchant was activated.
		var updatedMerchant models.Merchant

		if err := db.
			Where("id = ?", merchant.ID).
			First(&updatedMerchant).Error; err != nil {
			t.Fatalf("failed to fetch updated merchant: %v", err)
		}

		if updatedMerchant.Status != constants.MerchantStatusActive {
			t.Errorf(
				"expected merchant status %s, got %s",
				constants.MerchantStatusActive,
				updatedMerchant.Status,
			)
		}
	})
}

func TestSendWelcomeEmail(t *testing.T) {
	t.Run("returns SMTP error", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		user := &dto.CreateUserResponse{
			ID:        uuid.New(),
			Email:     "merchant@example.com",
			FirstName: "John",
			LastName:  "Doe",
		}

		temporaryPassword := "Temp@123"

		err := registry.SendWelcomeEmail(
			context.Background(),
			user,
			temporaryPassword,
		)

		if err == nil {
			t.Fatal("expected SMTP error, got nil")
		}
	})
}

func TestSendKYCReattemptEmail(t *testing.T) {
	t.Run("returns SMTP error", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		email := "merchant@example.com"
		merchantID := uuid.New()
		merchantType := constants.MerchantTypeIndividual

		err := registry.SendKYCReattemptEmail(
			context.Background(),
			email,
			merchantID,
			merchantType,
		)

		if err == nil {
			t.Fatal("expected SMTP error, got nil")
		}
	})
}

func TestUpdateMerchantCompliance(t *testing.T) {
	t.Run("merchant ID is required", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		complianceResponse := &dto.ComplianceCheckResponse{}

		merchant, err := registry.UpdateMerchantCompliance(
			context.Background(),
			uuid.Nil,
			complianceResponse,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if merchant != nil {
			t.Errorf("expected nil merchant, got %+v", merchant)
		}

		if err.Error() != "merchant_id is required" {
			t.Errorf("expected error %q, got %q", "merchant_id is required", err.Error())
		}
	})

	t.Run("compliance response is required", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchantID := uuid.New()

		merchant, err := registry.UpdateMerchantCompliance(
			context.Background(),
			merchantID,
			nil,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if merchant != nil {
			t.Errorf("expected nil merchant, got %+v", merchant)
		}

		if err.Error() != "compliance response is required" {
			t.Errorf("expected error %q, got %q", "compliance response is required", err.Error())
		}
	})

	t.Run("updates merchant compliance successfully", func(t *testing.T) {
		cleanTestDB(t)

		registry := NewRegistry(
			db,
			testServices.NotificationService,
			testServices,
			testRepo,
		)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		complianceDate := time.Now().Add(-time.Minute)
		kycDate := time.Now()

		complianceResponse := &dto.ComplianceCheckResponse{
			MerchantID:       merchant.ID,
			ComplianceStatus: constants.ComplianceStatusApproved,
			ComplianceDate:   complianceDate,
			ComplianceReason: "compliance_check_passed",
			KYCDate:          kycDate,
			KYC: dto.KYCData{
				Status: "approved",
				Reason: "Merchant passed all compliance checks.",
			},
			ComplianceDetails: dto.ComplianceDetails{},
			Country:           "India",
		}

		updatedMerchant, err := registry.UpdateMerchantCompliance(
			context.Background(),
			merchant.ID,
			complianceResponse,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if updatedMerchant == nil {
			t.Fatal("expected updated merchant, got nil")
		}

		if updatedMerchant.ID != merchant.ID {
			t.Errorf(
				"expected merchant ID %v, got %v",
				merchant.ID,
				updatedMerchant.ID,
			)
		}

		if updatedMerchant.ComplianceStatus != constants.ComplianceStatusApproved {
			t.Errorf(
				"expected compliance status %q, got %q",
				constants.ComplianceStatusApproved,
				updatedMerchant.ComplianceStatus,
			)
		}

		if updatedMerchant.ComplianceReason != "compliance_check_passed" {
			t.Errorf(
				"expected compliance reason %q, got %q",
				"compliance_check_passed",
				updatedMerchant.ComplianceReason,
			)
		}

		if updatedMerchant.Country != "India" {
			t.Errorf(
				"expected country %q, got %q",
				"India",
				updatedMerchant.Country,
			)
		}

		if updatedMerchant.ComplianceDate == nil {
			t.Fatal("expected compliance date, got nil")
		}

		if updatedMerchant.KYCDate == nil {
			t.Fatal("expected KYC date, got nil")
		}

		if len(updatedMerchant.KYC) == 0 {
			t.Fatal("expected KYC data, got empty JSON")
		}

		if len(updatedMerchant.ComplianceDetails) == 0 {
			t.Fatal("expected compliance details, got empty JSON")
		}

		var persistedMerchant models.Merchant

		if err := db.Where("id = ?", merchant.ID).First(&persistedMerchant).Error; err != nil {
			t.Fatalf("failed to fetch persisted merchant: %v", err)
		}

		if persistedMerchant.ComplianceStatus != constants.ComplianceStatusApproved {
			t.Errorf(
				"expected persisted compliance status %q, got %q",
				constants.ComplianceStatusApproved,
				persistedMerchant.ComplianceStatus,
			)
		}

		if persistedMerchant.ComplianceReason != "compliance_check_passed" {
			t.Errorf(
				"expected persisted compliance reason %q, got %q",
				"compliance_check_passed",
				persistedMerchant.ComplianceReason,
			)
		}

		if persistedMerchant.Country != "India" {
			t.Errorf(
				"expected persisted country %q, got %q",
				"India",
				persistedMerchant.Country,
			)
		}
	})
}
