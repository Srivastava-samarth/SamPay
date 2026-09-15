package controllers

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/constants"
	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

func (mc *Controller) CreateMerchant() gin.HandlerFunc {

	return func(c *gin.Context) {

		var request *dto.CreateMerchantOnboardingRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

		existingMerchant, errEM := mc.Services.GetMerchantByEmail(request.Email)
		if errEM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_FINDING_EXISTING_MERCHANT",
				errEM.Error(),
			)
		}

		if existingMerchant != nil {
			dto.Fail(
				c, http.StatusBadRequest,
				"MERCHANT_ALREADY_EXIST",
				"merchant already exist with same email",
			)
			return
		}

		if errV := mc.Services.ValidateMerchantOnboardingRequest(request); errV != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				errV.Error(),
			)
			return
		}

		// Create initial merchant here
		merchant, errM := mc.Services.CreateInitialMerchant(
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

func (mc *Controller) GetMerchantByID() gin.HandlerFunc {
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
		merchant, errM := mc.Services.GetMerchantByID(pasredMerchantId)
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

func (mc *Controller) GetMerchants() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchants, errM := mc.Services.GetMerchants()
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

func (mc *Controller) ReattemptOnboardingKyc() gin.HandlerFunc {
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

func (mc *Controller) UpdateMerchantInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		merchant, errM := mc.Services.GetMerchantByID(pasredMerchantId)
		if errM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"MERCHANT_NOT_FOUND",
				errM.Error(),
			)
			return
		}

		if merchant.Status != constants.MerchantStatusActive {
			dto.Fail(
				c,
				http.StatusConflict,
				"MERCHANT_NOT_ACTIVE",
				"merchant is not active",
			)
			return
		}

		updatedMerchant, errUM := mc.Services.UpdateMerchant(updateMerchantPayload, pasredMerchantId)
		if errUM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"MERCHANT_UPDATION_FAILED",
				errUM.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			updatedMerchant,
		)
	}
}

func (mc *Controller) UpdateMerchantStatus() gin.HandlerFunc {
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

		updatedMerchant, errUM := mc.Services.UpdateMerchantStatus(&request.Status, pasredMerchantId)
		if errUM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"UPDATION_MERCHANT_FAILED",
				errUM.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			updatedMerchant,
		)
	}
}

func (mc *Controller) UpdateMerchantKyc() gin.HandlerFunc {
	return func(c *gin.Context) {
		var updateMerchantPayload *dto.UpdateMerchantKycRequest

		if err := c.ShouldBindJSON(&updateMerchantPayload); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

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
				http.StatusBadRequest,
				"PARSING_ERROR",
				errPM.Error(),
			)
			return
		}

		updatedMerchant, errUM := mc.Services.UpdateMerchantKycInfo(updateMerchantPayload, pasredMerchantId)
		if errUM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"UPDATION_MERCHANT_FAILED",
				errUM.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			updatedMerchant,
		)
	}
}
