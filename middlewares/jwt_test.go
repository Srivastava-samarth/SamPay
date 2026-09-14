package middlewares

import (
	"testing"
	"time"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/google/uuid"
)

func testJWTConfig() *config.JWTConfig {
	secret := "test-secret"
	accessExpiry := "3600"
	refreshExpiry := "7200"

	return &config.JWTConfig{
		Secret:        secret,
		AccessExpiry:  accessExpiry,
		RefreshExpiry: refreshExpiry,
	}
}

func TestGenarateTokenAndExpiry(t *testing.T) {
	userID := uuid.New()
	merchantID := uuid.New()
	role := "owner"

	t.Run("valid token", func(t *testing.T) {
		j := NewJwt(testJWTConfig())

		token, err := j.GenarateTokenAndExpiry(
			userID,
			merchantID,
			role,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if token == "" {
			t.Fatal("expected token to be generated")
		}
	})

	t.Run("invalid access expiry", func(t *testing.T) {
		expiry := "invalid"

		cfg := testJWTConfig()
		cfg.AccessExpiry = expiry

		j := NewJwt(cfg)

		_, err := j.GenarateTokenAndExpiry(
			userID,
			merchantID,
			role,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGenerateRefreshToken(t *testing.T) {
	j := &Jwt{}

	token1, err := j.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token1 == "" {
		t.Fatal("expected refresh token to be generated")
	}

	token2, err := j.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token1 == token2 {
		t.Fatal("expected generated refresh tokens to be different")
	}
}

func TestGenerateResetPasswordToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"

	t.Run("valid token", func(t *testing.T) {
		j := NewJwt(testJWTConfig())

		token, err := j.GenerateResetPasswordToken(
			userID,
			email,
		)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if token == "" {
			t.Fatal("expected reset password token to be generated")
		}
	})

	t.Run("invalid access expiry", func(t *testing.T) {
		expiry := "invalid"

		cfg := testJWTConfig()
		cfg.AccessExpiry = expiry

		j := NewJwt(cfg)

		_, err := j.GenerateResetPasswordToken(
			userID,
			email,
		)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestValidateResetPaasword(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"

	j := NewJwt(testJWTConfig())

	validToken, err := j.GenerateResetPasswordToken(userID, email)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	tests := []struct {
		name          string
		token         string
		expectedError bool
	}{
		{
			name:          "valid token",
			token:         validToken,
			expectedError: false,
		},
		{
			name:          "malformed token",
			token:         "invalid-token",
			expectedError: true,
		},
		{
			name:          "empty token",
			token:         "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, gotEmail, err := j.ValidateResetPaasword(tt.token)

			if tt.expectedError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if gotUserID != userID {
				t.Fatalf("expected user ID %v, got %v", userID, gotUserID)
			}

			if gotEmail == "" || gotEmail != email {
				t.Fatalf("expected email %s, got %v", email, gotEmail)
			}
		})
	}

	t.Run("expired token", func(t *testing.T) {
		expiry := "-1"

		cfg := testJWTConfig()
		cfg.AccessExpiry = expiry

		expiredJwt := NewJwt(cfg)

		token, err := expiredJwt.GenerateResetPasswordToken(userID, email)
		if err != nil {
			t.Fatalf("failed to generate expired token: %v", err)
		}

		_, _, err = j.ValidateResetPaasword(token)

		if err == nil {
			t.Fatal("expected expired token to return error")
		}
	})
}

func TestValidateAccessToken(t *testing.T) {
	userID := uuid.New()
	merchantID := uuid.New()
	role := "owner"

	j := NewJwt(testJWTConfig())

	validToken, err := j.GenarateTokenAndExpiry(
		userID,
		merchantID,
		role,
	)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	tests := []struct {
		name          string
		token         string
		expectedError bool
	}{
		{
			name:          "valid token",
			token:         validToken,
			expectedError: false,
		},
		{
			name:          "malformed token",
			token:         "invalid-token",
			expectedError: true,
		},
		{
			name:          "empty token",
			token:         "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, gotMerchantID, gotRole, err :=
				j.ValidateAccessToken(tt.token)

			if tt.expectedError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if gotUserID != userID {
				t.Fatalf("expected user ID %v, got %v", userID, gotUserID)
			}

			if gotMerchantID != merchantID {
				t.Fatalf(
					"expected merchant ID %v, got %v",
					merchantID,
					gotMerchantID,
				)
			}

			if gotRole == "" || gotRole != role {
				t.Fatalf("expected role %s, got %v", role, gotRole)
			}
		})
	}

	t.Run("expired token", func(t *testing.T) {
		expiry := "-1"

		cfg := testJWTConfig()
		cfg.AccessExpiry = expiry

		expiredJwt := NewJwt(cfg)

		token, err := expiredJwt.GenarateTokenAndExpiry(
			userID,
			merchantID,
			role,
		)
		if err != nil {
			t.Fatalf("failed to generate expired token: %v", err)
		}

		_, _, _, err = j.ValidateAccessToken(token)

		if err == nil {
			t.Fatal("expected expired token to return error")
		}
	})

	t.Run("token with wrong secret", func(t *testing.T) {
		otherSecret := "different-secret"

		cfg := testJWTConfig()
		cfg.Secret = otherSecret

		otherJwt := NewJwt(cfg)

		token, err := otherJwt.GenarateTokenAndExpiry(
			userID,
			merchantID,
			role,
		)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		_, _, _, err = j.ValidateAccessToken(token)

		if err == nil {
			t.Fatal("expected error for token signed with different secret")
		}
	})

	_ = time.Now()
}
