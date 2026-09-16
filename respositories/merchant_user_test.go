package repositories

import (
	"errors"
	"testing"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestCreateMerchantUser(t *testing.T) {

	merchant := testutils.GenerateTestMerchant()
	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	email := uuid.NewString() + "@test.com"
	passwordHash := "test-password-hash"
	firstName := "Test"
	lastName := "User"
	status := "active"
	mustChangePassword := false

	user := &models.User{
		ID:                 uuid.New(),
		Email:              email,
		PasswordHash:       passwordHash,
		FirstName:          firstName,
		LastName:           lastName,
		Status:             status,
		MustChangePassword: mustChangePassword,
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	merchantUser := &models.MerchantUser{
		MerchantID: merchant.ID,
		UserID:     user.ID,
		Role:       "admin",
	}

	result, err := testRepo.CreateMerchantUser(merchantUser)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected merchant user, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected generated ID")
	}

	if result.MerchantID != merchant.ID {
		t.Errorf("expected merchant ID %v, got %v", merchant.ID, result.MerchantID)
	}

	if result.UserID != user.ID {
		t.Errorf("expected user ID %v, got %v", user.ID, result.UserID)
	}

	if result.Role != merchantUser.Role {
		t.Errorf("expected role %v, got %v", merchantUser.Role, result.Role)
	}

	if result.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if result.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestGetMerchantUserByUserID(t *testing.T) {

	merchant := testutils.GenerateTestMerchant()
	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	email := uuid.NewString() + "@test.com"
	passwordHash := "test-password-hash"
	firstName := "Test"
	lastName := "User"
	status := "active"
	mustChangePassword := false

	user := &models.User{
		ID:                 uuid.New(),
		Email:              email,
		PasswordHash:       passwordHash,
		FirstName:          firstName,
		LastName:           lastName,
		Status:             status,
		MustChangePassword: mustChangePassword,
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	merchantUser := &models.MerchantUser{
		MerchantID: merchant.ID,
		UserID:     user.ID,
		Role:       "admin",
	}

	if err := db.Create(merchantUser).Error; err != nil {
		t.Fatalf("failed to create merchant user: %v", err)
	}

	result, err := testRepo.GetMerchantUserByUserID(user.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected merchant user, got nil")
	}

	if result.UserID != user.ID {
		t.Errorf("expected user ID %v, got %v", user.ID, result.UserID)
	}

	if result.MerchantID != merchant.ID {
		t.Errorf("expected merchant ID %v, got %v", merchant.ID, result.MerchantID)
	}

	nonExistingUserID := uuid.New()

	_, err = testRepo.GetMerchantUserByUserID(nonExistingUserID)

	if err == nil {
		t.Fatal("expected error for non-existing user")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("expected record not found error, got %v", err)
	}
}

func TestGetMerchantUsersByMerchantID(t *testing.T) {

	merchant := testutils.GenerateTestMerchant()
	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("failed to create merchant: %v", err)
	}

	for i := 0; i < 2; i++ {
		email := uuid.NewString() + "@test.com"
		passwordHash := "test-password-hash"
		firstName := "Test"
		lastName := "User"
		status := "active"
		mustChangePassword := false

		user := &models.User{
			ID:                 uuid.New(),
			Email:              email,
			PasswordHash:       passwordHash,
			FirstName:          firstName,
			LastName:           lastName,
			Status:             status,
			MustChangePassword: mustChangePassword,
		}

		if err := db.Create(user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		merchantUser := &models.MerchantUser{
			ID:         uuid.New(),
			MerchantID: merchant.ID,
			UserID:     user.ID,
			Role:       "admin",
		}

		if err := db.Create(merchantUser).Error; err != nil {
			t.Fatalf("failed to create merchant user: %v", err)
		}
	}

	result, err := testRepo.GetMerchantUsersByMerchantID(merchant.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 merchant users, got %d", len(result))
	}

	nonExistingMerchantID := uuid.New()

	result, err = testRepo.GetMerchantUsersByMerchantID(nonExistingMerchantID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 merchant users, got %d", len(result))
	}
}
