package controllers

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/services"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"
)

type AuthController struct {
	TemporalClient client.Client
	AuthService    *services.AuthService
}

func NewAuthController(
	authService *services.AuthService,
	temporalClient client.Client,
) *AuthController {
	return &AuthController{
		AuthService:    authService,
		TemporalClient: temporalClient,
	}
}

func (ac *AuthController) Login() gin.HandlerFunc {
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

func (ac *AuthController) ForgotPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request *dto.ForgotPasswordRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

		workflowOptions := client.StartWorkflowOptions{
			ID:        "forgot password-" + request.Email,
			TaskQueue: temporal.ForgotPasswordTaskQueue,
		}

		_, err := ac.TemporalClient.ExecuteWorkflow(
			c.Request.Context(),
			workflowOptions,
			workflows.ForgotPasswordWorkflow,
			request,
		)

		if err != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"WORKFLOW_START_FAILED",
				err.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			"email sent to the user for reset password",
		)
	}
}

func (ac *AuthController) ResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request *dto.ResetPasswordRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

		err := ac.AuthService.ResetPassword(request)
		if err != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PASSWORD_RESET_FAILED",
				err.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			"Password Reset Successful",
		)
	}
}

func (ac *AuthController) RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request *dto.RefreshTokenRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

		response, errR := ac.AuthService.RefreshToken(request)
		if errR != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"TOKEN_REFRESH_FAILED",
				errR.Error(),
			)
			return
		}
		dto.Respond(
			c,
			http.StatusOK,
			response,
		)
	}
}
