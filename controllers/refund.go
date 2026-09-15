package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

func (rc *Controller) CreateRefund() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request *dto.CreateRefundRequest
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

		existingIdempotencyKey, errEI := rc.IdempotencyService.GetIdempotencyByKeyAndMerchantId(parsedMerchantID, idempotencyKey)
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

			rc.SaveIdempotencyResponse(
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
			ID:        "refund-flow-to-wallet" + merchantID,
			TaskQueue: temporal.RefundFlowTaskQueue,
		}

		workflowRun, errW := rc.TemporalClient.ExecuteWorkflow(
			c.Request.Context(),
			workflowOptions,
			workflows.RefundFlow,
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

			rc.SaveIdempotencyResponse(
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

		var refund *models.Refund

		err := workflowRun.Get(
			c.Request.Context(),
			&refund,
		)

		if err != nil {
			response := dto.Response{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "PAYOUT_FAILED",
					Message: err.Error(),
				},
			}

			rc.SaveIdempotencyResponse(
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
			Data:    refund,
		}

		rc.SaveIdempotencyResponse(
			parsedMerchantID,
			idempotencyKey,
			response,
		)

		dto.Respond(
			c,
			http.StatusOK,
			refund,
		)
	}
}

func (rc *Controller) GetRefundByReference() gin.HandlerFunc {
	return func(c *gin.Context) {
		refundRef := c.Param("refund_reference")
		if refundRef == "" {
			dto.Fail(
				c,
				http.StatusNotFound,
				"REFUND_REF_NOT_FOUND",
				"refund ref is wrong or not passed",
			)
			return
		}

		refund, errR := rc.RefundService.GetRefundByReference(refundRef)
		if errR != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ISSUE_GETTING_REFUND",
				errR.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			refund,
		)
	}
}

func (rc *Controller) GetRefundById() gin.HandlerFunc {
	return func(c *gin.Context) {
		refundId := c.Param("refund_id")
		if refundId == "" {
			dto.Fail(
				c,
				http.StatusNotFound,
				"REFUND_ID_NOT_FOUND",
				"refund id is wrong or not passed",
			)
			return
		}

		parsedRefundId, errP := uuid.Parse(refundId)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"REFUND_ID_PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		refund, errR := rc.RefundService.GetRefundById(parsedRefundId)
		if errR != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"REFUND_NOT_FOUND",
				errR.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			refund,
		)
	}
}

func (rc *Controller) GetRefundByMerchantId() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantId := c.Param("merchant_id")
		if merchantId == "" {
			dto.Fail(
				c,
				http.StatusNotFound,
				"MERCHANT_ID_NOT_FOUND",
				"Merchant ID not passed in params",
			)
			return
		}

		parsedMerchantId, errP := uuid.Parse(merchantId)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		refunds, errR := rc.RefundService.GetRefundsByMerchantId(parsedMerchantId)
		if errR != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"REFUNDS_NOT_FOUND",
				errR.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			gin.H{
				"data": refunds,
			},
		)
	}
}
