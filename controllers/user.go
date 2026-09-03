package controllers

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/services"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

type UserController struct {
	TemporalClient client.Client
	UserService *services.UserService
}

func NewUserController(
	temporalClient client.Client,
	userSrvc *services.UserService,
) *UserController {
	return &UserController{
		TemporalClient: temporalClient,
		UserService: userSrvc,
	}
}

func (uc *UserController) UserOnboarding() gin.HandlerFunc {
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
			ID: "user-onboarding-" + request.MerchantID.String(),
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

func (uc *UserController) GetUserByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exist := c.Get("user_id")
		if !exist{
			dto.Fail(
				c,
				http.StatusNotFound,
				"USER_ID",
				"User ID not found",
			)
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				"merchant id can't be parsed into string",
			)
			return
		}

		parsedUserId, errPM := uuid.Parse(userIDStr)
		if errPM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PARSING_ERROR",
				errPM.Error(),
			)
			return
		}
		user, errM := uc.UserService.GetUser(parsedUserId)
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

func (uc *UserController) GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchants, errM := uc.UserService.GetUsers()
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

func (uc *UserController) UpdateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exist := c.Get("user_id")
		if !exist {
			dto.Fail(
				c,
				http.StatusNotFound,
				"USER_ID_NOT_FOUND",
				"userId not found",
			)
			return
		}

		var updateUserPayload *dto.UpdateUserRequest
		if err := c.ShouldBindJSON(&updateUserPayload); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				"merchant id can't be parsed into string",
			)
			return
		}

		parsedUserId, errPM := uuid.Parse(userIDStr)
		if errPM != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				"error parsing merchant id in uuid",
			)
			return
		}
		updatedUser, errU := uc.UserService.UpdateUser(updateUserPayload, parsedUserId)
		if errU != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"USER_NOT_UPDATED",
				errU.Error(),
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
