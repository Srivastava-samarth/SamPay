package controllers

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/services"
	"github.com/gin-gonic/gin"
)

type AuthController struct{
	AuthService *services.AuthService
}

func NewAuthController(
	authService *services.AuthService,
) *AuthController{
	return &AuthController{
		AuthService: authService,
	}
}


func(ac *AuthController) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequest *dto.AuthRequest
		if err := c.ShouldBindJSON(&loginRequest); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

		response, err := ac.AuthService.Authentication(loginRequest)
		if err != nil {
			dto.Fail(
				c,
				http.StatusUnauthorized,
				"AUTHENTICATION_FAILED",
				err.Error(),
			)
			return
		}

		dto.Respond(c, http.StatusOK, response)
	}
}
