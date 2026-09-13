package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/services"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

type PayoutController struct {
	payoutSrvc      *services.PayoutService
	idempotencySrvc *services.IdempotencyService
	paymentCtrl     *PaymentController
	paymentSrvc     *services.PaymentService
	TemporalClient  client.Client
}

func NewPayoutController(
	payoutSrvc *services.PayoutService,
	idempotencySrvc *services.IdempotencyService,
	paymentCtrl *PaymentController,
	paymentSrvc *services.PaymentService,
	temporalClient client.Client,
) *PayoutController {
	return &PayoutController{
		payoutSrvc:      payoutSrvc,
		idempotencySrvc: idempotencySrvc,
		paymentCtrl:     paymentCtrl,
		paymentSrvc:     paymentSrvc,
		TemporalClient:  temporalClient,
	}
}

func (pc *PayoutController) WalletToBankAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request *dto.CreateWalletToBankRequest
		merchantID := c.Param("merchant_id")
		parsedMerchantID, errP := uuid.Parse(merchantID)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"MERCHANT_ID_PARSING_ERROR",
				errP.Error(),
			)
			return
		}
		idempotencyKey := strings.TrimSpace(
			c.GetHeader("Idempotency-Key"),
		)

		if idempotencyKey == "" {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"IDEMPOTENCY_KEY_NOT_FOUND",
				"Idempotency-Key header is required",
			)
			return
		}

		existingIdempotencyKey, errEI := pc.idempotencySrvc.GetIdempotencyByKeyAndMerchantId(parsedMerchantID, idempotencyKey)
		if errEI != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_FINDING_IDEMPOTENCY",
				errEI.Error(),
			)
			return
		}

		if existingIdempotencyKey != nil {

			var response dto.Response

			err := json.Unmarshal(
				existingIdempotencyKey.ResponseBody,
				&response,
			)
			if err != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"FAILED_UNMARSHALLING",
					err.Error(),
				)
				return
			}

			c.JSON(http.StatusOK, response)
			return
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			response := dto.Response{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "BINDING_ISSUE",
					Message: err.Error(),
				},
			}

			pc.paymentCtrl.SaveIdempotencyResponse(
				parsedMerchantID,
				idempotencyKey,
				response,
			)

			dto.Fail(
				c,
				http.StatusBadRequest,
				"BINDING_ISSUE",
				err.Error(),
			)
			return
		}

		workflowOptions := client.StartWorkflowOptions{
			ID:        "payout-flow-wallet-to-bank" + merchantID,
			TaskQueue: temporal.PayoutFlowTaskQueue,
		}

		workflowRun, errW := pc.TemporalClient.ExecuteWorkflow(
			c.Request.Context(),
			workflowOptions,
			workflows.PayoutFlow,
			request,
			parsedMerchantID,
		)

		if errW != nil {
			response := dto.Response{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "WORKFLOW_START_FAILED",
					Message: errW.Error(),
				},
			}

			pc.paymentCtrl.SaveIdempotencyResponse(
				parsedMerchantID,
				idempotencyKey,
				response,
			)

			dto.Fail(
				c,
				http.StatusInternalServerError,
				"WORKFLOW_START_FAILED",
				errW.Error(),
			)
			return
		}

		var payout *models.Payout

		err := workflowRun.Get(
			c.Request.Context(),
			&payout,
		)

		if err != nil {
			response := dto.Response{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "PAYOUT_FAILED",
					Message: err.Error(),
				},
			}

			pc.paymentCtrl.SaveIdempotencyResponse(
				parsedMerchantID,
				idempotencyKey,
				response,
			)

			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PAYOUT_FAILED",
				err.Error(),
			)
			return
		}

		response := dto.Response{
			Success: true,
			Data:    payout,
		}

		pc.paymentCtrl.SaveIdempotencyResponse(
			parsedMerchantID,
			idempotencyKey,
			response,
		)

		dto.Respond(
			c,
			http.StatusOK,
			payout,
		)
	}
}

func (pc *PayoutController) BankToBankAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request *dto.CreateBankToBankRequest
		merchantID := c.Param("merchant_id")
		parsedMerchantID, errP := uuid.Parse(merchantID)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"MERCHANT_ID_PARSING_ERROR",
				errP.Error(),
			)
			return
		}
		idempotencyKey := strings.TrimSpace(
			c.GetHeader("Idempotency-Key"),
		)

		if idempotencyKey == "" {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"IDEMPOTENCY_KEY_NOT_FOUND",
				"Idempotency-Key header is required",
			)
			return
		}

		existingIdempotencyKey, errEI := pc.idempotencySrvc.GetIdempotencyByKeyAndMerchantId(parsedMerchantID, idempotencyKey)
		if errEI != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_FINDING_IDEMPOTENCY",
				errEI.Error(),
			)
			return
		}

		if existingIdempotencyKey != nil {

			var response dto.Response

			err := json.Unmarshal(
				existingIdempotencyKey.ResponseBody,
				&response,
			)
			if err != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"FAILED_UNMARSHALLING",
					err.Error(),
				)
				return
			}

			c.JSON(http.StatusOK, response)
			return
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			response := dto.Response{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "BINDING_ISSUE",
					Message: err.Error(),
				},
			}

			pc.paymentCtrl.SaveIdempotencyResponse(
				parsedMerchantID,
				idempotencyKey,
				response,
			)

			dto.Fail(
				c,
				http.StatusBadRequest,
				"BINDING_ISSUE",
				err.Error(),
			)
			return
		}

		workflowOptions := client.StartWorkflowOptions{
			ID:        "payout-flow-bank-to-bank" + merchantID,
			TaskQueue: temporal.PayoutFlowTaskQueue,
		}

		workflowRun, errW := pc.TemporalClient.ExecuteWorkflow(
			c.Request.Context(),
			workflowOptions,
			workflows.PayoutFlowBankToBank,
			request,
			parsedMerchantID,
		)

		if errW != nil {
			response := dto.Response{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "WORKFLOW_START_FAILED",
					Message: errW.Error(),
				},
			}

			pc.paymentCtrl.SaveIdempotencyResponse(
				parsedMerchantID,
				idempotencyKey,
				response,
			)

			dto.Fail(
				c,
				http.StatusInternalServerError,
				"WORKFLOW_START_FAILED",
				errW.Error(),
			)
			return
		}

		var payout *models.Payout

		err := workflowRun.Get(
			c.Request.Context(),
			&payout,
		)

		if err != nil {
			response := dto.Response{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "PAYOUT_FAILED",
					Message: err.Error(),
				},
			}

			pc.paymentCtrl.SaveIdempotencyResponse(
				parsedMerchantID,
				idempotencyKey,
				response,
			)

			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PAYOUT_FAILED",
				err.Error(),
			)
			return
		}

		response := dto.Response{
			Success: true,
			Data:    payout,
		}

		pc.paymentCtrl.SaveIdempotencyResponse(
			parsedMerchantID,
			idempotencyKey,
			response,
		)

		dto.Respond(
			c,
			http.StatusOK,
			payout,
		)
	}
}
