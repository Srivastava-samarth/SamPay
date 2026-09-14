package repositories

import (
	"testing"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestCreateUserSession(t *testing.T) {
	repo := NewUserSessionRepository(db)

	refreshTokenHash := "test-refresh-token-" + uuid.NewString()
	expiresAt := time.Now().Add(time.Hour)

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

	session := &models.UserSession{
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        expiresAt,
	}

	result, err := repo.CreateUserSession(session)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected user session, got nil")
	}

	if result.ID == uuid.Nil {
		t.Error("expected generated ID")
	}

	if result.UserID != session.UserID {
		t.Errorf("expected user ID %v, got %v", session.UserID, result.UserID)
	}

	if result.RefreshTokenHash != refreshTokenHash {
		t.Errorf("expected refresh token hash %v, got %v", refreshTokenHash, result.RefreshTokenHash)
	}

	if !result.ExpiresAt.Equal(expiresAt) {
		t.Errorf("expected expiry %v, got %v", expiresAt, result.ExpiresAt)
	}

	if result.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if result.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestGetUserSessionByRefreshTokenHash(t *testing.T) {
	repo := NewUserSessionRepository(db)

	refreshTokenHash := "test-refresh-token-" + uuid.NewString()
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

	session := &models.UserSession{
		ID:               uuid.New(),
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        time.Now().Add(time.Hour),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.Create(session).Error; err != nil {
		t.Fatalf("failed to create user session: %v", err)
	}

	result, err := repo.GetUserSessionByRefreshTokenHash(refreshTokenHash)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected user session, got nil")
	}

	if result.ID != session.ID {
		t.Errorf("expected session ID %v, got %v", session.ID, result.ID)
	}

	if result.RefreshTokenHash != refreshTokenHash {
		t.Errorf("expected refresh token hash %v, got %v", refreshTokenHash, result.RefreshTokenHash)
	}

	invalidHash := "invalid-refresh-token-" + uuid.NewString()

	_, err = repo.GetUserSessionByRefreshTokenHash(invalidHash)

	if err == nil {
		t.Fatal("expected error for invalid refresh token hash")
	}

	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected record not found error, got %v", err)
	}
}
