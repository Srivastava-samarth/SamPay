package controllers

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/dto"
	services "github.com/Srivastava-samarth/sampay/services"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func(mc *MerchantController) GetMerchantByID() gin.HandlerFunc{
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

		pasredMerchantId, errPM := uuid.Parse(merchantID)
		if errPM != nil{
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PARSING_ERROR",
				errPM.Error(),
			)
			return
		}
		merchant, errM := mc.MerchantService.GetMerchantByID(pasredMerchantId)
		if errM != nil{
			dto.Fail(
				c,
				http.StatusNotFound,
				"MERCHANT_NOT_FOUND",
				errPM.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			merchant,
		)
	}
}

func(mc *MerchantController) GetMerchants() gin.HandlerFunc{
	return func(c *gin.Context) {
		merchants, errM := mc.MerchantService.GetMerchants()
		if errM != nil{
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"MERCHANTS_NOT_FOUND",
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
