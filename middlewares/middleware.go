package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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