package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

var logger = utils.NewLogger()

func (pc *Controller) CreatePayment() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request *dto.CreatePaymentRequest
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

		existingIdempotencyKey, errEI := pc.Services.GetIdempotencyByKeyAndMerchantId(parsedMerchantID, idempotencyKey)
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

			pc.SaveIdempotencyResponse(
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
			ID:        "payment-flow-" + merchantID,
			TaskQueue: temporal.PaymentFlowTaskQueue,
		}

		workflowRun, errW := pc.TemporalClient.ExecuteWorkflow(
			c.Request.Context(),
			workflowOptions,
			workflows.PaymentFlow,
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

			pc.SaveIdempotencyResponse(
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

		var payment *models.Payment

		err := workflowRun.Get(
			c.Request.Context(),
			&payment,
		)

		if err != nil {
			response := dto.Response{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "PAYMENT_FAILED",
					Message: err.Error(),
				},
			}

			pc.SaveIdempotencyResponse(
				parsedMerchantID,
				idempotencyKey,
				response,
			)

			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PAYMENT_FAILED",
				err.Error(),
			)
			return
		}

		response := dto.Response{
			Success: true,
			Data:    payment,
		}

		pc.SaveIdempotencyResponse(
			parsedMerchantID,
			idempotencyKey,
			response,
		)

		dto.Respond(
			c,
			http.StatusOK,
			payment,
		)
	}
}

func (pc *Controller) SaveIdempotencyResponse(
	merchantID uuid.UUID,
	idempotencyKey string,
	response dto.Response,
) {
	responseBody, err := json.Marshal(response)
	if err != nil {
		logger.Print("Error marshaling idempotency response: ", err)
		return
	}

	_, err = pc.Services.CreateIdempotencyKey(
		merchantID,
		idempotencyKey,
		responseBody,
	)
	if err != nil {
		logger.Print("Error saving idempotency response: ", err)
	}
}

func (pc *Controller) TriggerSettlement() gin.HandlerFunc {
	return func(c *gin.Context) {
		workflowOptions := client.StartWorkflowOptions{
			ID: fmt.Sprintf(
				"settlement-%s",
				time.Now().UTC().Format("20060102-150405"),
			),
			TaskQueue: temporal.SettlementFlowTaskQueue,
		}

		_, errW := pc.TemporalClient.ExecuteWorkflow(
			c.Request.Context(),
			workflowOptions,
			workflows.SettlementFlow,
		)

		if errW != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"SETTLEMENT_FAILED",
				errW.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			workflowOptions.ID,
		)
	}
}

func (pc *Controller) GetPaymentsByMerchantID() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantID := c.Param("merchant_id")
		if merchantID == "" {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"MERCHANT_ID_NOT_FOUND",
				"Merchant ID not passed in params",
			)
		}

		parsedMerchantID, errP := uuid.Parse(merchantID)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		payments, errGP := pc.Services.GetPaymentsByMerchantID(parsedMerchantID)
		if errGP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_FINDING_PAYMENTS",
				errGP.Error(),
			)
		}

		dto.Respond(
			c,
			http.StatusOK,
			payments,
		)
	}
}

func (pc *Controller) GetPaymentByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		paymentID := c.Param("payment_id")
		if paymentID == "" {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"PAYMENT_ID_NOT_FOUND",
				"Payment ID not passed in params",
			)
		}

		parsedPaymentID, errP := uuid.Parse(paymentID)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		payment, errGP := pc.Services.GetPaymentByID(parsedPaymentID)
		if errGP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_GETTING_PAYMENT",
				errGP.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			payment,
		)
	}
}
