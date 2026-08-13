package controllers

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/dto"
	services "github.com/Srivastava-samarth/sampay/services"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"
)

type MerchantController struct{
	MerchantService *services.MerchantService
	TemporalClient client.Client
}

func NewMerchantController(
	merchantSrvc *services.MerchantService,
	temporalClient client.Client,
) *MerchantController{
	return &MerchantController{
		MerchantService: merchantSrvc,
		TemporalClient: temporalClient,
	}
}


func(mc *MerchantController) CreateMerchant() gin.HandlerFunc {

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
		merchant, err := mc.MerchantService.CreateInitialMerchant(
			request,
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

		_, err = mc.TemporalClient.ExecuteWorkflow(
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
