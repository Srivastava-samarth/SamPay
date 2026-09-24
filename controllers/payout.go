package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

func (pc *Controller) WalletToBankAccount() gin.HandlerFunc {
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

		// Try to claim the idempotency key.
		createdIdempotencyKey, errCI :=
			pc.Services.CreateIdempotencyKey(
				parsedMerchantID,
				idempotencyKey,
			)

		if errCI != nil {
			if errors.Is(errCI, gorm.ErrDuplicatedKey) {
				existingIdempotencyKey, errEI :=
					pc.Services.GetIdempotencyByKeyAndMerchantId(
						parsedMerchantID,
						idempotencyKey,
					)

				if errEI != nil {
					dto.Fail(
						c,
						http.StatusInternalServerError,
						"ERROR_FINDING_IDEMPOTENCY",
						errEI.Error(),
					)
					return
				}

				if existingIdempotencyKey == nil {
					dto.Fail(
						c,
						http.StatusInternalServerError,
						"IDEMPOTENCY_NOT_FOUND",
						"idempotency key was created by another request but could not be found",
					)
					return
				}

				// The original request is still processing.
				// We will replace this with waiting for the original
				// workflow in the next step.
				if existingIdempotencyKey.ResponseBody == nil {
					dto.Fail(
						c,
						http.StatusConflict,
						"REQUEST_IN_PROGRESS",
						"request with this idempotency key is already in progress",
					)
					return
				}

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

			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_CREATING_IDEMPOTENCY",
				errCI.Error(),
			)
			return
		}

		// We successfully claimed the idempotency key.
		if err := c.ShouldBindJSON(&request); err != nil {
			response := dto.Response{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "BINDING_ISSUE",
					Message: err.Error(),
				},
			}

			responseBody, errM := json.Marshal(response)
			if errM != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"FAILED_MARSHALLING",
					errM.Error(),
				)
				return
			}

			createdIdempotencyKey.ResponseBody = responseBody

			_, errU := pc.Services.UpdateIdempotencyKey(
				createdIdempotencyKey,
			)
			if errU != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"ERROR_UPDATING_IDEMPOTENCY",
					errU.Error(),
				)
				return
			}

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

			responseBody, errM := json.Marshal(response)
			if errM != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"FAILED_MARSHALLING",
					errM.Error(),
				)
				return
			}

			createdIdempotencyKey.ResponseBody = responseBody

			_, errU := pc.Services.UpdateIdempotencyKey(
				createdIdempotencyKey,
			)
			if errU != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"ERROR_UPDATING_IDEMPOTENCY",
					errU.Error(),
				)
				return
			}

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

			responseBody, errM := json.Marshal(response)
			if errM != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"FAILED_MARSHALLING",
					errM.Error(),
				)
				return
			}

			createdIdempotencyKey.ResponseBody = responseBody

			_, errU := pc.Services.UpdateIdempotencyKey(
				createdIdempotencyKey,
			)
			if errU != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"ERROR_UPDATING_IDEMPOTENCY",
					errU.Error(),
				)
				return
			}

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

		responseBody, errM := json.Marshal(response)
		if errM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"FAILED_MARSHALLING",
				errM.Error(),
			)
			return
		}

		createdIdempotencyKey.ResponseBody = responseBody

		_, errU := pc.Services.UpdateIdempotencyKey(
			createdIdempotencyKey,
		)
		if errU != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_UPDATING_IDEMPOTENCY",
				errU.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			payout,
		)
	}
}

func (pc *Controller) BankToBankAccount() gin.HandlerFunc {
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

		createdIdempotencyKey, errCI :=
			pc.Services.CreateIdempotencyKey(
				parsedMerchantID,
				idempotencyKey,
			)

		if errCI != nil {
			if errors.Is(errCI, gorm.ErrDuplicatedKey) {
				existingIdempotencyKey, errEI :=
					pc.Services.GetIdempotencyByKeyAndMerchantId(
						parsedMerchantID,
						idempotencyKey,
					)

				if errEI != nil {
					dto.Fail(
						c,
						http.StatusInternalServerError,
						"ERROR_FINDING_IDEMPOTENCY",
						errEI.Error(),
					)
					return
				}

				if existingIdempotencyKey == nil {
					dto.Fail(
						c,
						http.StatusInternalServerError,
						"IDEMPOTENCY_NOT_FOUND",
						"idempotency key was created by another request but could not be found",
					)
					return
				}

				if existingIdempotencyKey.ResponseBody == nil {
					dto.Fail(
						c,
						http.StatusConflict,
						"REQUEST_IN_PROGRESS",
						"request with this idempotency key is already in progress",
					)
					return
				}

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

			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_CREATING_IDEMPOTENCY",
				errCI.Error(),
			)
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

			responseBody, errM := json.Marshal(response)
			if errM != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"FAILED_MARSHALLING",
					errM.Error(),
				)
				return
			}

			createdIdempotencyKey.ResponseBody = responseBody

			_, errU := pc.Services.UpdateIdempotencyKey(
				createdIdempotencyKey,
			)
			if errU != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"ERROR_UPDATING_IDEMPOTENCY",
					errU.Error(),
				)
				return
			}

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

			responseBody, errM := json.Marshal(response)
			if errM != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"FAILED_MARSHALLING",
					errM.Error(),
				)
				return
			}

			createdIdempotencyKey.ResponseBody = responseBody

			_, errU := pc.Services.UpdateIdempotencyKey(
				createdIdempotencyKey,
			)
			if errU != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"ERROR_UPDATING_IDEMPOTENCY",
					errU.Error(),
				)
				return
			}

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

			responseBody, errM := json.Marshal(response)
			if errM != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"FAILED_MARSHALLING",
					errM.Error(),
				)
				return
			}

			createdIdempotencyKey.ResponseBody = responseBody

			_, errU := pc.Services.UpdateIdempotencyKey(
				createdIdempotencyKey,
			)
			if errU != nil {
				dto.Fail(
					c,
					http.StatusInternalServerError,
					"ERROR_UPDATING_IDEMPOTENCY",
					errU.Error(),
				)
				return
			}

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

		responseBody, errM := json.Marshal(response)
		if errM != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"FAILED_MARSHALLING",
				errM.Error(),
			)
			return
		}

		createdIdempotencyKey.ResponseBody = responseBody

		_, errU := pc.Services.UpdateIdempotencyKey(
			createdIdempotencyKey,
		)
		if errU != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_UPDATING_IDEMPOTENCY",
				errU.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			payout,
		)
	}
}

func (pc *Controller) GetPayoutsByMerchantID() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantID := c.Param("merchant_id")
		parsedMerchantID, errP := uuid.Parse(merchantID)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		payouts, errGP := pc.Services.GetPayoutsByMerchantID(parsedMerchantID)
		if errGP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ISSUE_GETTING_PAYOUTS",
				errGP.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			payouts,
		)
	}
}

func (pc *Controller) GetPayoutByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		payoutID := c.Param("payout_id")
		parsedPayoutID, errP := uuid.Parse(payoutID)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		payout, errGP := pc.Services.GetPayoutByID(parsedPayoutID)
		if errGP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ISSUE_GETTING_PAYOUT",
				errGP.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			payout,
		)
	}
}
