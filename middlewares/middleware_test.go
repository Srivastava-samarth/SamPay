package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func middlewareJWTConfig() *config.JWTConfig {
	secret := "test-secret"
	accessExpiry := "3600"
	refreshExpiry := "7200"

	return &config.JWTConfig{
		Secret:        secret,
		AccessExpiry:  accessExpiry,
		RefreshExpiry: refreshExpiry,
	}
}

func TestAuthenticate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	merchantID := uuid.New()
	role := "owner"

	j := NewJwt(middlewareJWTConfig())

	validToken, err := j.GenarateTokenAndExpiry(
		userID,
		merchantID,
		role,
	)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tests := []struct {
		name               string
		authorization      string
		expectedStatusCode int
		expectedAborted    bool
	}{
		{
			name:               "missing authorization header",
			authorization:      "",
			expectedStatusCode: http.StatusUnauthorized,
			expectedAborted:    true,
		},
		{
			name:               "invalid authorization header",
			authorization:      "InvalidToken",
			expectedStatusCode: http.StatusUnauthorized,
			expectedAborted:    true,
		},
		{
			name:               "invalid bearer format",
			authorization:      "Basic " + validToken,
			expectedStatusCode: http.StatusUnauthorized,
			expectedAborted:    true,
		},
		{
			name:               "invalid access token",
			authorization:      "Bearer invalid-token",
			expectedStatusCode: http.StatusUnauthorized,
			expectedAborted:    true,
		},
		{
			name:               "valid access token",
			authorization:      "Bearer " + validToken,
			expectedStatusCode: http.StatusOK,
			expectedAborted:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			router.Use(j.Authenticate())

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "success",
				})
			})

			req := httptest.NewRequest(
				http.MethodGet,
				"/test",
				nil,
			)

			if tt.authorization != "" {
				req.Header.Set(
					"Authorization",
					tt.authorization,
				)
			}

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.expectedStatusCode {
				t.Fatalf(
					"expected status %d, got %d",
					tt.expectedStatusCode,
					recorder.Code,
				)
			}

			if tt.expectedAborted {
				if recorder.Code == http.StatusOK {
					t.Fatal("expected request to be aborted")
				}
			}
		})
	}

	t.Run("valid token sets context values", func(t *testing.T) {
		router := gin.New()

		router.Use(j.Authenticate())

		router.GET("/test", func(c *gin.Context) {
			gotUserID, exists := c.Get("user_id")
			if !exists {
				t.Fatal("expected user_id in context")
			}

			gotMerchantID, exists := c.Get("merchant_id")
			if !exists {
				t.Fatal("expected merchant_id in context")
			}

			gotRole, exists := c.Get("role")
			if !exists {
				t.Fatal("expected role in context")
			}

			if gotUserID != userID {
				t.Fatalf(
					"expected user ID %v, got %v",
					userID,
					gotUserID,
				)
			}

			if gotMerchantID != merchantID {
				t.Fatalf(
					"expected merchant ID %v, got %v",
					merchantID,
					gotMerchantID,
				)
			}

			if gotRole != role {
				t.Fatalf(
					"expected role %s, got %s",
					role,
					gotRole,
				)
			}

			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(
			http.MethodGet,
			"/test",
			nil,
		)

		req.Header.Set(
			"Authorization",
			"Bearer "+validToken,
		)

		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusOK,
				recorder.Code,
			)
		}
	})
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adminRole := "admin"
	userRole := "user"
	merchantRole := "merchant"

	tests := []struct {
		name               string
		contextRole        interface{}
		allowedRoles       []string
		expectedStatusCode int
	}{
		{
			name:               "role not found",
			contextRole:        nil,
			allowedRoles:       []string{adminRole},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "invalid role type",
			contextRole:        "admin",
			allowedRoles:       []string{"basdj"},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:               "role allowed",
			contextRole:        adminRole,
			allowedRoles:       []string{adminRole},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "role allowed among multiple roles",
			contextRole:        userRole,
			allowedRoles:       []string{adminRole, userRole},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "role not allowed",
			contextRole:        &userRole,
			allowedRoles:       []string{adminRole, merchantRole},
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			router.Use(func(c *gin.Context) {
				if tt.contextRole != nil {
					c.Set("role", tt.contextRole)
				}

				c.Next()
			})

			router.Use(RequireRole(tt.allowedRoles...))

			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(
				http.MethodGet,
				"/test",
				nil,
			)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.expectedStatusCode {
				t.Fatalf(
					"expected status %d, got %d",
					tt.expectedStatusCode,
					recorder.Code,
				)
			}
		})
	}
}

func TestRequireMerchantAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	merchantRole := "merchant"

	tokenMerchantID := uuid.New()
	differentMerchantID := uuid.New()

	tests := []struct {
		name               string
		role               interface{}
		tokenMerchantID    interface{}
		requestedMerchant  string
		expectedStatusCode int
	}{
		{
			name:               "role not found",
			role:               nil,
			tokenMerchantID:    tokenMerchantID,
			requestedMerchant:  tokenMerchantID.String(),
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "invalid merchant identity type",
			role:               &merchantRole,
			tokenMerchantID:    "invalid-merchant-id",
			requestedMerchant:  tokenMerchantID.String(),
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "invalid requested merchant id",
			role:               &merchantRole,
			tokenMerchantID:    tokenMerchantID,
			requestedMerchant:  "invalid-uuid",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "same merchant access",
			role:               &merchantRole,
			tokenMerchantID:    tokenMerchantID,
			requestedMerchant:  tokenMerchantID.String(),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "different merchant denied",
			role:               &merchantRole,
			tokenMerchantID:    tokenMerchantID,
			requestedMerchant:  differentMerchantID.String(),
			expectedStatusCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			router.Use(func(c *gin.Context) {
				if tt.role != nil {
					c.Set("role", tt.role)
				}

				if tt.tokenMerchantID != nil {
					c.Set("merchant_id", tt.tokenMerchantID)
				}

				c.Next()
			})

			router.Use(RequireMerchantAccess(merchantRole))

			router.GET(
				"/merchant/:merchant_id",
				func(c *gin.Context) {
					c.Status(http.StatusOK)
				},
			)

			req := httptest.NewRequest(
				http.MethodGet,
				"/merchant/"+tt.requestedMerchant,
				nil,
			)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.expectedStatusCode {
				t.Fatalf(
					"expected status %d, got %d",
					tt.expectedStatusCode,
					recorder.Code,
				)
			}
		})
	}
}
