package controllers

import (
	"net/http"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	services "github.com/Srivastava-samarth/sampay/services"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

type MerchantController struct {
	MerchantService *services.MerchantService
	TemporalClient  client.Client
}

func NewMerchantController(
	merchantSrvc *services.MerchantService,
	temporalClient client.Client,
) *MerchantController {
	return &MerchantController{
		MerchantService: merchantSrvc,
		TemporalClient:  temporalClient,
	}
}

func (mc *MerchantController) CreateMerchant() gin.HandlerFunc {

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

		if errV := mc.MerchantService.ValidateMerchantOnboardingRequest(&request); errV != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				errV.Error(),
			)
			return
		}

		// Create initial merchant here
		merchant, errM := mc.MerchantService.CreateInitialMerchant(
			request,
		)

		if errM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"MERCHANT_CREATION_FAILED",
				errM.Error(),
			)
			return
		}

		workflowOptions := client.StartWorkflowOptions{
			ID:        "merchant-onboarding-" + merchant.ID.String(),
			TaskQueue: temporal.MerchantOnboardingTaskQueue,
		}

		_, errW := mc.TemporalClient.ExecuteWorkflow(
			c.Request.Context(),
			workflowOptions,
			workflows.MerchantONboardingWorkflow,
			request,
			merchant.ID,
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

		dto.Respond(
			c,
			http.StatusAccepted,
			merchant,
		)
	}
}

func (mc *MerchantController) GetMerchantByID() gin.HandlerFunc {
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
		if errPM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PARSING_ERROR",
				errPM.Error(),
			)
			return
		}
		merchant, errM := mc.MerchantService.GetMerchantByID(pasredMerchantId)
		if errM != nil {
			dto.Fail(
				c,
				http.StatusNotFound,
				"MERCHANT_NOT_FOUND",
				errM.Error(),
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

func (mc *MerchantController) GetMerchants() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchants, errM := mc.MerchantService.GetMerchants()
		if errM != nil {
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

func (mc *MerchantController) UpdateMerchant() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantID, exist := c.Get("merchant_id")
		if !exist {
			dto.Fail(
				c,
				http.StatusNotFound,
				"MERCHANT_ID_NOT_FOUND",
				"merchantId not found",
			)
			return
		}

		var updateMerchantPayload *dto.UpdateMerchantRequest
		if err := c.ShouldBindJSON(&updateMerchantPayload); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

		merchantIDStr, ok := merchantID.(string)
		if !ok {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				"merchant id can't be parsed into string",
			)
			return
		}

		parsedMerchantId, errPM := uuid.Parse(merchantIDStr)
		if errPM != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				"error parsing merchant id in uuid",
			)
			return
		}
		merchant, errM := mc.MerchantService.GetMerchantByID(parsedMerchantId)
		if errM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"MERCHANT_NOT_FOUND",
				errM.Error(),
			)
			return
		}

		var (
			updatedMerchant *models.Merchant
			errUM           error
		)
		if merchant.MerchantType == "individual" {
			updatedMerchant, errUM = mc.MerchantService.UpdateIndividualMerchant(updateMerchantPayload, merchant)
			if errUM != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"MERCHANT_UPDATION_FAILED",
					errUM.Error(),
				)
				return
			}
		} else {
			updatedMerchant, errUM = mc.MerchantService.UpdateCompanyMerchant(updateMerchantPayload, merchant)
			if errUM != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"MERCHANT_UPDATION_FAILED",
					errUM.Error(),
				)
				return
			}
		}
		dto.Respond(
			c,
			http.StatusOK,
			updatedMerchant,
		)
	}
}

func (mc *MerchantController) UpdateKYC() gin.HandlerFunc{
	return func(c *gin.Context) {
		var request *dto.UpdateKYCRequest
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
		if errPM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PARSING_ERROR",
				errPM.Error(),
			)
			return
		}

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
			ID:        "merchant-update-kyc-" + merchantID,
			TaskQueue: temporal.UpdateKycFlow,
		}

		workflowRun, errW := mc.TemporalClient.ExecuteWorkflow(
			c.Request.Context(),
			workflowOptions,
			workflows.MerchantUpdateKycFlow,
			request,
			pasredMerchantId,
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

		var merchant *models.Merchant

		errM := workflowRun.Get(
			c.Request.Context(),
			&merchant,
		)

		if errM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"FAILED_TO_EXTRACT_MERCHANT",
				errM.Error(),
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
