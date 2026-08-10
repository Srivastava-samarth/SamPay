package controllers

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/dto"
	services "github.com/Srivastava-samarth/sampay/services"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

var validate = validator.New()

func CreateMerchant(
	db *gorm.DB,
	temporalClient client.Client,
) gin.HandlerFunc {

	return func(c *gin.Context) {

		var request dto.CreateMerchantOnboardingRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

		// Create initial merchant here
		merchant, err := services.CreateInitialMerchant(
			request,
			db,
		)

		if err != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"MERCHANT_CREATION_FAILED",
				err.Error(),
			)
			return
		}

		workflowOptions := client.StartWorkflowOptions{
			ID: "merchant-onboarding-" + merchant.ID.String(),
			TaskQueue: temporal.MerchantOnboardingTaskQueue,
		}

		_, err = temporalClient.ExecuteWorkflow(
			c.Request.Context(),
			workflowOptions,
			workflows.MerchantONboardingWorkflow,
			request,
			merchant.ID,
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
			http.StatusAccepted,
			merchant,
		)
	}
}
