package controllers

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"
)

type UserController struct {
	TemporalClient client.Client
}

func NewUserController(
	temporalClient client.Client,
) *UserController {
	return &UserController{
		TemporalClient: temporalClient,
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
