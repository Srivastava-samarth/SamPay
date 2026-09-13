package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (j *Jwt) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "authorization header required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid authorization header",
			})
			return
		}

		userID, merchantID, role, err :=
			j.ValidateAccessToken(parts[1])

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid or expired access token",
			})
			return
		}

		c.Set("user_id", userID)
		c.Set("merchant_id", merchantID)
		c.Set("role", role)

		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {

		roleValue, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "user role not found",
			})
			return
		}

		role, ok := roleValue.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid user role",
			})
			return
		}

		for _, allowedRole := range roles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "insufficient permissions",
		})
	}
}

func RequireMerchantAccess(role string) gin.HandlerFunc {
	return func(c *gin.Context) {

		roleAssigned, roleExist := c.Get("role")
		if !roleExist {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "role not found",
			})
			return
		}

		if roleAssigned == role {
			c.Next()
		}

		merchantIDValue, exists := c.Get("merchant_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "merchant identity not found",
			})
			return
		}

		tokenMerchantID, ok := merchantIDValue.(uuid.UUID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid merchant identity",
			})
			return
		}

		requestedMerchantID, err := uuid.Parse(
			c.Param("merchant_id"),
		)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "invalid merchant id",
			})
			return
		}

		if tokenMerchantID != requestedMerchantID && roleAssigned != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "access denied",
			})
			return
		}

		c.Next()
	}
}
