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
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

var logger = utils.NewLogger()

type PaymentController struct {
	db              *gorm.DB
	paymentSrvc     *services.PaymentService
	idempotencySrvc *services.IdempotencyService
	walletSrvc      *services.WalletService
	vaultSrvc       *services.VaultService
	ledgerSrvc      *services.LedgerService
	TemporalClient  client.Client
}

func NewPaymentController(
	db *gorm.DB,
	paymentSrvc *services.PaymentService,
	idempotencySrvc *services.IdempotencyService,
	walletSrvc *services.WalletService,
	vaultSrvc *services.VaultService,
	ledgerSrvc *services.LedgerService,
	temporalClient *client.Client,
) *PaymentController {
	return &PaymentController{
		db:              db,
		paymentSrvc:     paymentSrvc,
		idempotencySrvc: idempotencySrvc,
		walletSrvc:      walletSrvc,
		vaultSrvc:       vaultSrvc,
		ledgerSrvc:      ledgerSrvc,
		TemporalClient:  *temporalClient,
	}
}

func (pc *PaymentController) CreatePayment() gin.HandlerFunc {
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

func (pc *PaymentController) SaveIdempotencyResponse(
	merchantID uuid.UUID,
	idempotencyKey string,
	response dto.Response,
) {
	responseBody, err := json.Marshal(response)
	if err != nil {
		logger.Print("Error marshaling idempotency response: ", err)
		return
	}

	_, err = pc.idempotencySrvc.CreateIdempotencyKey(
		merchantID,
		idempotencyKey,
		responseBody,
	)
	if err != nil {
		logger.Print("Error saving idempotency response: ", err)
	}
}
