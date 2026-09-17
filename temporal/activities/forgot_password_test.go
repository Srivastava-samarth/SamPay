package activities

import (
	"context"
	"testing"

	"github.com/Srivastava-samarth/sampay/config"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/notifications"
	"github.com/Srivastava-samarth/sampay/services"
	"github.com/google/uuid"
)

func TestForgotPassword(t *testing.T) {
	t.Run("user not found", func(t *testing.T) {
		cleanTestDB(t)

		notificationService, err := notifications.NewEmailService(
			config.SMTPConfig{},
		)
		if err != nil {
			t.Fatalf("failed to create notification service: %v", err)
		}

		services := services.NewServices(
			db,
			testRepo,
			testServices.JwtService,
			notificationService,
		)

		registry := NewRegistry(
			db,
			notificationService,
			services,
			testRepo,
		)

		email := "does-not-exist@example.com"

		request := &dto.ForgotPasswordRequest{
			Email: email,
		}

		errR := registry.ForgotPassword(
			context.Background(),
			request,
		)

		if errR == nil {
			t.Fatal("expected error, got nil")
		}

		if errR == nil {
			t.Fatal("expected underlying error, got nil")
		}
	})

	t.Run("valid user - notification failure", func(t *testing.T) {
		cleanTestDB(t)

		user := &models.User{
			ID:    uuid.New(),
			Email: "test@example.com",
		}

		if err := db.Create(user).Error; err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}

		notificationService, err := notifications.NewEmailService(
			config.SMTPConfig{},
		)
		if err != nil {
			t.Fatalf("failed to create notification service: %v", err)
		}

		services := services.NewServices(
			db,
			testRepo,
			testServices.JwtService,
			notificationService,
		)

		registry := NewRegistry(
			db,
			notificationService,
			services,
			testRepo,
		)

		request := &dto.ForgotPasswordRequest{
			Email: user.Email,
		}

		errR := registry.ForgotPassword(
			context.Background(),
			request,
		)

		if errR == nil {
			t.Fatal("expected notification error, got nil")
		}

		if errR == nil {
			t.Fatal("expected underlying error, got nil")
		}

		var resetToken models.PasswordResetToken

		errDB := db.
			Where("user_id = ?", user.ID).
			First(&resetToken).Error

		if errDB != nil {
			t.Fatalf(
				"expected password reset token to be created before notification failure, got: %v",
				errDB,
			)
		}

		if resetToken.TokenHash == "" {
			t.Error("expected reset token hash to be populated")
		}
	})
}
