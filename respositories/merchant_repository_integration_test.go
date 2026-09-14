package repositories

import (
	"encoding/json"
	"testing"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/google/uuid"
)

func TestCreateMerchant(t *testing.T) {
	repo := NewMerchantRepository(db)

	name := "Test Merchant"
	email := uuid.NewString() + "@test.com"
	phone := "9999999999"
	status := "active"
	merchantType := "individual"
	complianceStatus := "pending"

	merchant := &models.Merchant{
		MerchantName:     name,
		Email:            email,
		PhoneNumber:      phone,
		Status:           status,
		MerchantType:     merchantType,
		ComplianceStatus: complianceStatus,
	}

	result, err := repo.CreateMerchant(*merchant)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected merchant, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected generated ID")
	}

	if result.MerchantReference == "" {
		t.Error("expected generated merchant reference")
	}

	if result.MerchantName != merchant.MerchantName {
		t.Error("expected merchant name to match")
	}

	if result.Email != merchant.Email {
		t.Error("expected email to match")
	}
}

func TestGetMerchantByID(t *testing.T) {
	repo := NewMerchantRepository(db)

	merchant := testutils.GenerateTestMerchant()

	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	result, err := repo.GetMerchantByID(merchant.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected merchant, got nil")
	}

	if result.ID != merchant.ID {
		t.Errorf("expected merchant ID %v, got %v", merchant.ID, result.ID)
	}

	nonExistingID := uuid.New()

	result, err = repo.GetMerchantByID(nonExistingID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Error("expected nil merchant for non-existing ID")
	}
}

func TestGetMerchants(t *testing.T) {
	repo := NewMerchantRepository(db)

	merchant1 := testutils.GenerateTestMerchant()
	merchant2 := testutils.GenerateTestMerchant()

	if err := db.Create(merchant1).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	if err := db.Create(merchant2).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	result, err := repo.GetMerchants()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) < 2 {
		t.Fatalf("expected at least 2 merchants, got %d", len(result))
	}

	found1, found2 := false, false

	for _, merchant := range result {
		if merchant.ID == merchant1.ID {
			found1 = true
		}

		if merchant.ID == merchant2.ID {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("expected created merchants to be returned")
	}
}

func TestUpdateMerchantCompliance(t *testing.T) {
	repo := NewMerchantRepository(db)

	merchant := testutils.GenerateTestMerchant()

	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	complianceStatus := "approved"
	complianceReason := "KYC verified"
	country := "IN"
	kycStatus := "verified"
	kycReason := "documents verified"

	complianceResponse := &dto.ComplianceCheckResponse{
		ComplianceStatus: complianceStatus,
		ComplianceReason: complianceReason,
		Country:          country,
		KYCDate:          time.Now(),
		ComplianceDate:   time.Now(),
		KYC: dto.KYCData{
			Status: kycStatus,
			Reason: kycReason,
		},
		ComplianceDetails: dto.ComplianceDetails{},
	}

	result, err := repo.UpdateMerchantCompliance(merchant.ID, complianceResponse)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected merchant, got nil")
	}

	if result.ComplianceStatus != complianceStatus {
		t.Errorf("expected compliance status %v, got %v", complianceStatus, result.ComplianceStatus)
	}

	if result.ComplianceReason != complianceReason {
		t.Errorf("expected compliance reason %v, got %v", complianceReason, result.ComplianceReason)
	}

	if result.Country != country {
		t.Errorf("expected country %v, got %v", country, result.Country)
	}

	var kyc dto.KYCData
	if err := json.Unmarshal(result.KYC, &kyc); err != nil {
		t.Fatalf("failed to unmarshal KYC: %v", err)
	}

	if kyc.Status != kycStatus {
		t.Errorf("expected KYC status %v, got %v", kycStatus, kyc.Status)
	}

	if kyc.Reason != kycReason {
		t.Errorf("expected KYC reason %v, got %v", kycReason, kyc.Reason)
	}
}

func TestUpdateMerchantStatus(t *testing.T) {
	repo := NewMerchantRepository(db)

	merchant := testutils.GenerateTestMerchant()

	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	status := "inactive"

	result, err := repo.UpdateMerchantStatus(merchant.ID, status)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected merchant, got nil")
	}

	if result.Status != status {
		t.Errorf("expected status %v, got %v", status, result.Status)
	}

	if result.ID != merchant.ID {
		t.Errorf("expected merchant ID %v, got %v", merchant.ID, result.ID)
	}
}

func TestUpdateMerchant(t *testing.T) {
	repo := NewMerchantRepository(db)

	merchant := testutils.GenerateTestMerchant()

	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	newName := "Updated Merchant"
	newPhone := "8888888888"

	request := &dto.UpdateMerchantRequest{
		MerchantName: newName,
		PhoneNumber:  newPhone,
	}

	result, err := repo.UpdateMerchant(request, merchant.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected merchant, got nil")
	}

	if result.MerchantName != newName {
		t.Errorf("expected merchant name %v, got %v", newName, result.MerchantName)
	}

	if result.PhoneNumber != newPhone {
		t.Errorf("expected phone number %v, got %v", newPhone, result.PhoneNumber)
	}

	oldName := result.MerchantName

	request = &dto.UpdateMerchantRequest{
		MerchantName: "",
		PhoneNumber:  "",
	}

	result, err = repo.UpdateMerchant(request, merchant.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.MerchantName != oldName {
		t.Errorf("expected merchant name to remain %v, got %v", oldName, result.MerchantName)
	}

	if result.PhoneNumber != newPhone {
		t.Errorf("expected phone number to remain %v, got %v", newPhone, result.PhoneNumber)
	}
}

func TestGetMerchantByEmail(t *testing.T) {
	repo := NewMerchantRepository(db)

	merchant := testutils.GenerateTestMerchant()

	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	result, err := repo.GetMerchantByEmail(merchant.Email)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected merchant, got nil")
	}

	if result.ID != merchant.ID {
		t.Errorf("expected merchant ID %v, got %v", merchant.ID, result.ID)
	}

	if result.Email == "" || result.Email != merchant.Email {
		t.Errorf("expected email %v, got %v", merchant.Email, result.Email)
	}

	nonExistingEmail := uuid.NewString() + "@test.com"

	result, err = repo.GetMerchantByEmail(nonExistingEmail)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Error("expected nil merchant for non-existing email")
	}
}
