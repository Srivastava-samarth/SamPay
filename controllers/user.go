package controllers

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

func (uc *Controller) UserOnboarding() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request dto.UserOnboardingRequest
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
			ID:        "user-onboarding-" + request.MerchantID.String(),
			TaskQueue: temporal.UserOnboardingTaskQueue,
		}

		_, errW := uc.TemporalClient.ExecuteWorkflow(
			c.Request.Context(),
			workflowOptions,
			workflows.UserOnboardingFlow,
			request,
		)

		if errW != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"WORKFLOW_START_FAILED",
				errW.Error(),
			)
			return
		}

		dto.Respond(c, http.StatusOK, request.Email)
	}
}

func (uc *Controller) GetUserByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")

		parsedUserId, errPM := uuid.Parse(userID)
		if errPM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PARSING_ERROR",
				errPM.Error(),
			)
			return
		}
		user, errM := uc.Services.GetUser(parsedUserId)
		if errM != nil {
			dto.Fail(
				c,
				http.StatusNotFound,
				"USER_NOT_FOUND",
				errM.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			user,
		)
	}
}

func (uc *Controller) GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchants, errM := uc.Services.GetUsers()
		if errM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"USERS_NOT_FOUND",
				errM.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			merchants,
		)
	}
}

func (uc *Controller) UpdateUserStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request *dto.UpdateMerchantStatusRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

		userID := c.Param("user_id")
		parsedUserID, errP := uuid.Parse(userID)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		updatedUser, errUU := uc.Services.UpdateUserStatus(request.Status, parsedUserID)
		if errUU != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"UPDATE_USER_ISSUE",
				errUU.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			updatedUser,
		)
	}
}

func (uc *Controller) GetUsersByMerchant() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantID := c.Param("merchant_id")
		if merchantID == "" {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"MERCHANT_ID_NOT_FOUND",
				"Merchant ID not passed in params",
			)
			return
		}

		parsedMerchantID, errP := uuid.Parse(merchantID)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		users, errU := uc.Services.GetUsersByMerchant(parsedMerchantID)
		if errU != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"GETTING_USERS_ISSUE",
				errU.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			users,
		)
	}
}
