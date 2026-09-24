package activities

import (
	"context"
	"testing"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/notifications"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
)

func TestCreateUser(t *testing.T) {
	activityRegistry := &Registry{
		DB:       db,
		Services: testServices,
		Repo:     testRepo,
	}
	t.Run("request is nil", func(t *testing.T) {
		result, err := testServices.CreateUser(nil)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "request is required" {
			t.Errorf(
				"expected error %q, got %q",
				"request is required",
				(err).Error(),
			)
		}

		if result != nil {
			t.Error("expected nil result, got non-nil")
		}
	})

	t.Run("email is required", func(t *testing.T) {
		request := &dto.CreateUserRequest{
			FirstName: "Sam",
			LastName:  "Test",
		}

		result, err := testServices.CreateUser(request)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "email is required" {
			t.Errorf(
				"expected error %q, got %q",
				"email is required",
				(err).Error(),
			)
		}

		if result != nil {
			t.Error("expected nil result, got non-nil")
		}
	})

	t.Run("first name is required", func(t *testing.T) {
		request := &dto.CreateUserRequest{
			Email:    "sam@example.com",
			LastName: "Test",
		}

		result, err := testServices.CreateUser(request)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "first name is required" {
			t.Errorf(
				"expected error %q, got %q",
				"first name is required",
				(err).Error(),
			)
		}

		if result != nil {
			t.Error("expected nil result, got non-nil")
		}
	})

	t.Run("last name is required", func(t *testing.T) {
		request := &dto.CreateUserRequest{
			Email:     "sam@example.com",
			FirstName: "Sam",
		}

		result, err := testServices.CreateUser(request)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "last name is required" {
			t.Errorf(
				"expected error %q, got %q",
				"last name is required",
				(err).Error(),
			)
		}

		if result != nil {
			t.Error("expected nil result, got non-nil")
		}
	})

	t.Run("creates user successfully", func(t *testing.T) {
		request := &dto.CreateUserRequest{
			Email:     "sam-create-user@example.com",
			FirstName: "Sam",
			LastName:  "Test",
		}

		result, err := activityRegistry.CreateUser(context.Background(), request)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		if result.User == nil {
			t.Fatal("expected user response, got nil")
		}

		if result.TemporaryPassword == "" {
			t.Error("expected temporary password to be generated")
		}

		if len(result.TemporaryPassword) != 8 {
			t.Errorf(
				"expected temporary password length 8, got %d",
				len(result.TemporaryPassword),
			)
		}

		if result.User.ID == uuid.Nil {
			t.Error("expected user ID to be generated")
		}

		if result.User.Email != request.Email {
			t.Errorf(
				"expected email %q, got %q",
				request.Email,
				result.User.Email,
			)
		}

		if result.User.FirstName != request.FirstName {
			t.Errorf(
				"expected first name %q, got %q",
				request.FirstName,
				result.User.FirstName,
			)
		}

		if result.User.LastName != request.LastName {
			t.Errorf(
				"expected last name %q, got %q",
				request.LastName,
				result.User.LastName,
			)
		}

		var storedUser models.User

		if err := db.
			Where("id = ?", result.User.ID).
			First(&storedUser).Error; err != nil {
			t.Fatalf("failed to fetch created user: %v", err)
		}

		if storedUser.ID != result.User.ID {
			t.Errorf(
				"expected stored user ID %v, got %v",
				result.User.ID,
				storedUser.ID,
			)
		}

		if storedUser.Email != request.Email {
			t.Errorf(
				"expected stored email %q, got %q",
				request.Email,
				storedUser.Email,
			)
		}

		if storedUser.FirstName != request.FirstName {
			t.Errorf(
				"expected stored first name %q, got %q",
				request.FirstName,
				storedUser.FirstName,
			)
		}

		if storedUser.LastName != request.LastName {
			t.Errorf(
				"expected stored last name %q, got %q",
				request.LastName,
				storedUser.LastName,
			)
		}

		if storedUser.PasswordHash == "" {
			t.Fatal("expected password hash to be stored")
		}

		if storedUser.PasswordHash == result.TemporaryPassword {
			t.Error("password must not be stored as plaintext")
		}
	})
}

func TestCreateMerchantUser(t *testing.T) {
	activityRegistry := &Registry{
		DB:       db,
		Services: testServices,
		Repo:     testRepo,
	}
	t.Run("request is nil", func(t *testing.T) {
		result, err := testServices.CreateMerchantUser(nil)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "request is required" {
			t.Errorf(
				"expected error %q, got %q",
				"request is required",
				(err).Error(),
			)
		}

		if result != nil {
			t.Error("expected nil result, got non-nil")
		}
	})

	t.Run("merchant id is required", func(t *testing.T) {
		request := &dto.CreateMerchantUserRequest{
			UserID: uuid.New(),
			Role:   "admin",
		}

		result, err := testServices.CreateMerchantUser(request)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "merchant_id is required" {
			t.Errorf(
				"expected error %q, got %q",
				"merchant_id is required",
				(err).Error(),
			)
		}

		if result != nil {
			t.Error("expected nil result, got non-nil")
		}
	})

	t.Run("user id is required", func(t *testing.T) {
		request := &dto.CreateMerchantUserRequest{
			MerchantID: uuid.New(),
			Role:       "admin",
		}

		result, err := testServices.CreateMerchantUser(request)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "user_id is required" {
			t.Errorf(
				"expected error %q, got %q",
				"user_id is required",
				(err).Error(),
			)
		}

		if result != nil {
			t.Error("expected nil result, got non-nil")
		}
	})

	t.Run("role is required", func(t *testing.T) {
		request := &dto.CreateMerchantUserRequest{
			MerchantID: uuid.New(),
			UserID:     uuid.New(),
		}

		result, err := testServices.CreateMerchantUser(request)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if (err).Error() != "role is required" {
			t.Errorf(
				"expected error %q, got %q",
				"role is required",
				(err).Error(),
			)
		}

		if result != nil {
			t.Error("expected nil result, got non-nil")
		}
	})

	t.Run("creates merchant user successfully", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()
		user := &models.User{
			ID:                 utils.GenerateUUID(),
			Email:              "test.merchant@sampay.io",
			PasswordHash:       "$2a$10$pnhmHrwWKUdy1Kh5CHHNo.3Pv4YhjEYwzGCxreJ/WAwvYJGRBkUWW",
			Status:             constants.UserStatusActive,
			MustChangePassword: false,
			FirstName:          "Test",
			LastName:           "Merchant",
		}

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		if err := db.Create(user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		request := &dto.CreateMerchantUserRequest{
			MerchantID: merchant.ID,
			UserID:     user.ID,
			Role:       "admin",
		}

		result, err := activityRegistry.CreateMerchantUser(context.Background(), request)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		if result.MerchantUserLinkedID == uuid.Nil {
			t.Error("expected merchant user linked ID to be generated")
		}

		if result.MerchantID != merchant.ID {
			t.Errorf(
				"expected merchant ID %v, got %v",
				merchant.ID,
				result.MerchantID,
			)
		}

		if result.UserID != user.ID {
			t.Errorf(
				"expected user ID %v, got %v",
				user.ID,
				result.UserID,
			)
		}

		if result.Role != request.Role {
			t.Errorf(
				"expected role %q, got %q",
				request.Role,
				result.Role,
			)
		}

		var storedMerchantUser models.MerchantUser

		if err := db.
			Where("id = ?", result.MerchantUserLinkedID).
			First(&storedMerchantUser).Error; err != nil {
			t.Fatalf("failed to fetch created merchant user: %v", err)
		}

		if storedMerchantUser.MerchantID != merchant.ID {
			t.Errorf(
				"expected stored merchant ID %v, got %v",
				merchant.ID,
				storedMerchantUser.MerchantID,
			)
		}

		if storedMerchantUser.UserID != user.ID {
			t.Errorf(
				"expected stored user ID %v, got %v",
				user.ID,
				storedMerchantUser.UserID,
			)
		}

		if storedMerchantUser.Role != request.Role {
			t.Errorf(
				"expected stored role %q, got %q",
				request.Role,
				storedMerchantUser.Role,
			)
		}
	})
}

func TestGetMerchantById(t *testing.T) {
	activityRegistry := &Registry{
		DB:       db,
		Services: testServices,
		Repo:     testRepo,
	}
	t.Run("returns merchant successfully", func(t *testing.T) {
		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		result, err := activityRegistry.GetMerchantById(merchant.ID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("expected merchant, got nil")
		}

		if result.ID != merchant.ID {
			t.Errorf(
				"expected merchant ID %v, got %v",
				merchant.ID,
				result.ID,
			)
		}
	})

	t.Run("merchant does not exist", func(t *testing.T) {
		merchantID := uuid.New()

		result, err := testRepo.GetMerchantByID(merchantID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result != nil {
			t.Errorf("expected nil merchant, got %v", result)
		}
	})
}

func TestSendUserWelcomeEmail(t *testing.T) {
	activityRegistry := &Registry{
		DB:       db,
		Services: testServices,
		Repo:     testRepo,
		NotificationService: &notifications.EmailService{
			Config: config.SMTPConfig{
				Host: "127.0.0.1",
				Port: "1",
				From: "noreply@sampay.com",
			},
		},
	}
	t.Run("returns error when SMTP server is unavailable", func(t *testing.T) {
		name := "Test User"
		email := "test@example.com"
		merchantName := "Test Merchant"
		role := "admin"
		temporaryPassword := "Temp1234"

		err := activityRegistry.SendUserWelcomeEmail(
			context.Background(),
			name,
			email,
			merchantName,
			role,
			temporaryPassword,
		)

		if err == nil {
			t.Fatalf("expected SMTP error, got nil")
		}
	})
}
