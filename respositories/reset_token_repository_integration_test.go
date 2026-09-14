package repositories

import (
	"testing"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestCreatePasswordReset(t *testing.T) {
	repo := NewPasswordResetTokenRepository(db)

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
	tokenHash := "test-token-hash"
	expiresAt := time.Now().Add(time.Hour)
	createdAt := time.Now()

	resetToken := &models.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
	}

	result, err := repo.CreatePasswordReset(resetToken)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected reset token, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected generated ID")
	}

	if result.UserID != user.ID {
		t.Errorf("expected user ID %v, got %v", user.ID, result.UserID)
	}

	if result.TokenHash != tokenHash {
		t.Errorf("expected token hash %v, got %v", tokenHash, result.TokenHash)
	}

	if !result.ExpiresAt.Equal(expiresAt) {
		t.Errorf("expected expiry %v, got %v", expiresAt, result.ExpiresAt)
	}
}

func TestUpdatePasswordReset(t *testing.T) {
	repo := NewPasswordResetTokenRepository(db)

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
	resetToken := &models.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: "test-token-hash-" + uuid.NewString(),
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	if err := db.Create(resetToken).Error; err != nil {
		t.Fatalf("failed to create reset token: %v", err)
	}

	usedAt := time.Now()
	resetToken.UsedAt = usedAt

	result, err := repo.UpdatePasswordReset(resetToken)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected reset token, got nil")
	}

	if result.ID != resetToken.ID {
		t.Errorf("expected ID %v, got %v", resetToken.ID, result.ID)
	}

	nonExisting := &models.PasswordResetToken{
		ID:     uuid.New(),
		UsedAt: usedAt,
	}

	_, err = repo.UpdatePasswordReset(nonExisting)

	if err == nil {
		t.Fatal("expected record not found error")
	}

	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected record not found error, got %v", err)
	}
}

func TestFindByToken(t *testing.T) {
	repo := NewPasswordResetTokenRepository(db)

	token := "test-reset-token-" + uuid.NewString()
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

	resetToken := &models.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: utils.HashToken(token),
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	if err := db.Create(resetToken).Error; err != nil {
		t.Fatalf("failed to create reset token: %v", err)
	}

	result, err := repo.FindByToken(token)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected reset token, got nil")
	}

	if result.ID != resetToken.ID {
		t.Errorf("expected ID %v, got %v", resetToken.ID, result.ID)
	}

	if result.TokenHash != resetToken.TokenHash {
		t.Errorf("expected token hash %v, got %v", resetToken.TokenHash, result.TokenHash)
	}

	invalidToken := "invalid-token-" + uuid.NewString()

	_, err = repo.FindByToken(invalidToken)

	if err == nil {
		t.Fatal("expected error for invalid token")
	}

	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected record not found error, got %v", err)
	}
}
