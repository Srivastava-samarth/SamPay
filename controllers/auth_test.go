package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/testutils"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid request", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST("/login", testController.Login())

		req := httptest.NewRequest(
			http.MethodPost,
			"/login",
			bytes.NewBufferString(`{"email":`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}
	})

	t.Run("authentication failed", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST("/login", testController.Login())

		body := `{
			"email": "nonexistent@test.com",
			"password": "password"
		}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/login",
			bytes.NewBufferString(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf(
				"expected status %d, got %d",
				http.StatusUnauthorized,
				rec.Code,
			)
		}
	})

	t.Run("successful login", func(t *testing.T) {
		cleanTestDB(t)

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		email := uuid.NewString() + "@test.com"
		password := "test-password"
		passwordHash, err := bcrypt.GenerateFromPassword(
			[]byte(password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}

		firstName := "Test"
		lastName := "User"
		status := constants.UserStatusActive
		mustChangePassword := false

		user := &models.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: string(passwordHash),
			FirstName:    firstName,
			LastName:     lastName,
			Status:       status,
		}

		if err := db.Create(user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		userUpdate := &dto.UpdateUserRequest{
			MustChangePassword: &mustChangePassword,
		}

		newUser, errUU := testRepo.UpdateUser(user.ID, userUpdate)
		if errUU != nil {
			t.Fatalf("failed to update user: %v", err)
		}
		// Create merchant-user relationship.
		role := "owner"

		merchantUser := &models.MerchantUser{
			ID:         uuid.New(),
			UserID:     user.ID,
			MerchantID: merchant.ID,
			Role:       role,
		}

		if err := db.Create(merchantUser).Error; err != nil {
			t.Fatalf("failed to create merchant user: %v", err)
		}

		var persistedUser models.User

		if err := db.Where("id = ?", newUser.ID).First(&persistedUser).Error; err != nil {
			t.Fatalf("failed to fetch user: %v", err)
		}

		if persistedUser.MustChangePassword {
			t.Fatal("expected MustChangePassword to be false")
		}

		router := gin.New()
		router.POST("/login", testController.Login())

		body := fmt.Sprintf(`{
			"email": "%s",
			"password": "%s"
		}`, email, password)

		req := httptest.NewRequest(
			http.MethodPost,
			"/login",
			bytes.NewBufferString(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		if rec.Body.Len() == 0 {
			t.Fatal("expected response body")
		}
		var session models.UserSession
		if err := db.
			Where("user_id = ?", user.ID).
			First(&session).Error; err != nil {
			t.Fatalf("expected user session to be created: %v", err)
		}

		if session.RefreshTokenHash == "" {
			t.Error("expected refresh token hash to be created")
		}

		if session.ExpiresAt.IsZero() {
			t.Error("expected refresh token expiry to be set")
		}
	})
}

func TestForgotPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid request", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST("/forgot-password", testController.ForgotPassword())

		req := httptest.NewRequest(
			http.MethodPost,
			"/forgot-password",
			bytes.NewBufferString(`{"email":`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("workflow started successfully", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST("/forgot-password", testController.ForgotPassword())

		body := `{"email":"nonexistent@test.com"}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/forgot-password",
			bytes.NewBufferString(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		expectedMessage := "email sent to the user for reset password"
		if !strings.Contains(rec.Body.String(), expectedMessage) {
			t.Fatalf(
				"expected response to contain %q, got %s",
				expectedMessage,
				rec.Body.String(),
			)
		}
	})

}

func TestResetPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid request", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST("/reset-password", testController.ResetPassword())

		req := httptest.NewRequest(
			http.MethodPost,
			"/reset-password",
			bytes.NewBufferString(`{"email":`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}
	})

	t.Run("password reset failed", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST("/reset-password", testController.ResetPassword())

		body := `{
			"email": "nonexistent@test.com",
			"reset_token": "invalid-reset-token",
			"new_password": "new-password",
			"confirm_password": "new-password"
		}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/reset-password",
			bytes.NewBufferString(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})

	t.Run("successful password reset", func(t *testing.T) {
		cleanTestDB(t)

		email := uuid.NewString() + "@test.com"
		oldPassword := "old-password"
		firstName := "Test"
		lastName := "User"
		status := constants.UserStatusActive
		mustChangePassword := true

		oldPasswordHash, err := utils.HashPassword(oldPassword)
		if err != nil {
			t.Fatalf("failed to hash old password: %v", err)
		}

		user := &models.User{
			ID:                 uuid.New(),
			Email:              email,
			PasswordHash:       oldPasswordHash,
			FirstName:          firstName,
			LastName:           lastName,
			Status:             status,
			MustChangePassword: mustChangePassword,
		}

		if err := db.Create(user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		resetToken, err := testServices.JwtService.GenerateResetPasswordToken(
			user.ID,
			user.Email,
		)
		if err != nil {
			t.Fatalf("failed to generate reset token: %v", err)
		}

		hashedResetToken := utils.HashToken(resetToken)

		passwordResetToken := &models.PasswordResetToken{
			ID:        uuid.New(),
			UserID:    user.ID,
			TokenHash: hashedResetToken,
			ExpiresAt: time.Now().Add(time.Hour),
			CreatedAt: time.Now(),
		}

		if err := db.Create(passwordResetToken).Error; err != nil {
			t.Fatalf("failed to create password reset token: %v", err)
		}

		router := gin.New()
		router.POST("/reset-password", testController.ResetPassword())

		newPassword := "new-password"

		body := fmt.Sprintf(`{
		"email": "%s",
		"reset_token": "%s",
		"new_password": "%s",
		"confirm_password": "%s"
	}`,
			email,
			resetToken,
			newPassword,
			newPassword,
		)

		req := httptest.NewRequest(
			http.MethodPost,
			"/reset-password",
			bytes.NewBufferString(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		if !strings.Contains(
			rec.Body.String(),
			"Password Reset Successful",
		) {
			t.Fatalf(
				"expected successful response, got %s",
				rec.Body.String(),
			)
		}

		updatedUser := &models.User{}
		if err := db.First(updatedUser, "id = ?", user.ID).Error; err != nil {
			t.Fatalf("failed to fetch updated user: %v", err)
		}

		if bcrypt.CompareHashAndPassword(
			[]byte(updatedUser.PasswordHash),
			[]byte(newPassword),
		) != nil {
			t.Fatal("expected user password to be updated")
		}

		if updatedUser.MustChangePassword {
			t.Fatal("expected MustChangePassword to be false")
		}

		updatedResetToken := &models.PasswordResetToken{}
		if err := db.First(
			updatedResetToken,
			"id = ?",
			passwordResetToken.ID,
		).Error; err != nil {
			t.Fatalf("failed to fetch updated reset token: %v", err)
		}

		if updatedResetToken.UsedAt.IsZero() {
			t.Fatal("expected reset token to be marked as used")
		}
	})

}

func TestRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid request", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST("/refresh-token", testController.RefreshToken())

		req := httptest.NewRequest(
			http.MethodPost,
			"/refresh-token",
			bytes.NewBufferString(`{"refresh_token":`),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusBadRequest,
				rec.Code,
			)
		}
	})

	t.Run("token refresh failed", func(t *testing.T) {
		cleanTestDB(t)

		router := gin.New()
		router.POST("/refresh-token", testController.RefreshToken())

		body := `{
			"refresh_token": "invalid-refresh-token"
		}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/refresh-token",
			bytes.NewBufferString(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusInternalServerError,
				rec.Code,
			)
		}
	})

	t.Run("successful token refresh", func(t *testing.T) {
		cleanTestDB(t)

		email := uuid.NewString() + "@test.com"
		passwordHash := "test-password-hash"
		firstName := "Test"
		lastName := "User"
		status := constants.UserStatusActive
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

		merchant := testutils.GenerateTestMerchant()

		if err := db.Create(merchant).Error; err != nil {
			t.Fatalf("failed to create merchant: %v", err)
		}

		role := "owner"

		merchantUser := &models.MerchantUser{
			ID:         uuid.New(),
			UserID:     user.ID,
			MerchantID: merchant.ID,
			Role:       role,
		}

		if err := db.Create(merchantUser).Error; err != nil {
			t.Fatalf("failed to create merchant user: %v", err)
		}

		refreshToken := "test-refresh-token"
		refreshTokenHash := utils.HashToken(refreshToken)

		userSession := &models.UserSession{
			ID:               uuid.New(),
			UserID:           user.ID,
			RefreshTokenHash: refreshTokenHash,
			ExpiresAt:        time.Now().Add(time.Hour),
		}

		if err := db.Create(userSession).Error; err != nil {
			t.Fatalf("failed to create user session: %v", err)
		}

		router := gin.New()
		router.POST("/refresh-token", testController.RefreshToken())

		body := fmt.Sprintf(`{
		"refresh_token": "%s"
	}`, refreshTokenHash)

		req := httptest.NewRequest(
			http.MethodPost,
			"/refresh-token",
			bytes.NewBufferString(body),
		)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d, body: %s",
				http.StatusOK,
				rec.Code,
				rec.Body.String(),
			)
		}

		var response struct {
			Success bool `json:"success"`
			Data    struct {
				AccessToken *string `json:"access_token"`
			} `json:"data"`
		}

		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !response.Success {
			t.Fatal("expected successful response")
		}

		if response.Data.AccessToken == nil || *response.Data.AccessToken == "" {
			t.Fatal("expected access token in response")
		}
	})

}
