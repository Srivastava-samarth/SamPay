package controllers

import (
	"net/http"

	models "github.com/Srivastava-samarth/sampay/database/models"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/workflows"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

func (c *Controller) GetBankAccount() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		bankAccountID := ctx.Param("bank_account_id")

		parsedBankAccountID, errP := uuid.Parse(bankAccountID)
		if errP != nil{
			dto.Fail(
				ctx,
				http.StatusBadRequest,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		bankAccount, errBA := c.Services.GetBankAccount(parsedBankAccountID)
		if errBA != nil{
			dto.Fail(
				ctx,
				http.StatusInternalServerError,
				"ERROR_FETCHING_BANK_ACCOUNT",
				errBA.Error(),
			)
			return
		}

		dto.Respond(
			ctx,
			http.StatusOK,
			bankAccount,
		)
	}
}

func (c *Controller) GetBankAccounts() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		merchantID := ctx.Param("merchant_id")

		parsedMerchantID, errP := uuid.Parse(merchantID)
		if errP != nil{
			dto.Fail(
				ctx,
				http.StatusBadRequest,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		bankAccounts, errBA := c.Services.GetBankAccountsByMerchantID(parsedMerchantID)
		if errBA != nil{
			dto.Fail(
				ctx,
				http.StatusInternalServerError,
				"ERROR_FETCHING_BANK_ACCOUNT",
				errBA.Error(),
			)
			return
		}

		dto.Respond(
			ctx,
			http.StatusOK,
			bankAccounts,
		)
	}
}

func (c *Controller) CreateBankAccount() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		merchantID := ctx.Param("merchant_id")
		parsedMerchantID, errP := uuid.Parse(merchantID)
		if errP != nil{
			dto.Fail(
				ctx,
				http.StatusBadRequest,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		var request *dto.CreateBankAccountRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			dto.Fail(
				ctx,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

		if errV := c.Services.ValidateCreateBankAccountRequest(ctx, request, parsedMerchantID);  errV != nil{
			dto.Fail(
				ctx,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				errV.Error(),
			)
			return
		}

		workflowOptions := client.StartWorkflowOptions{
			ID:        "bank-account-" + merchantID,
			TaskQueue: temporal.BankAccountTaskQueue,
		}

		workflowRun, errW := c.TemporalClient.ExecuteWorkflow(
			ctx.Request.Context(),
			workflowOptions,
			workflows.CreateBankAccount,
			request,
			parsedMerchantID,
		)

		if errW != nil {
			dto.Fail(
				ctx,
				http.StatusInternalServerError,
				"WORKFLOW_START_FAILED",
				errW.Error(),
			)
			return
		}

		var bankAccount *models.BankAccount

		err := workflowRun.Get(
			ctx.Request.Context(),
			&bankAccount,
		)

		if err != nil{
			dto.Fail(
				ctx,
				http.StatusInternalServerError,
				"ERROR_POPULATING_BANK_ACCOUNT",
				err.Error(),
			)
		}

		dto.Respond(
			ctx,
			http.StatusOK,
			bankAccount,
		)
	}
}

func (c *Controller) UpdateBankAccountStatus() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		var request *dto.UpdateBankAccountStatus
		if err := ctx.ShouldBindJSON(&request); err!= nil{
			dto.Fail(
				ctx,
				http.StatusBadRequest,
				"BINDING_ERROR",
				err.Error(),
			)
			return
		}

		bankAccountID := ctx.Param("bank_account_id")
		parsedBankAccountID, errP := uuid.Parse(bankAccountID)
		if errP != nil{
			dto.Fail(
				ctx,
				http.StatusBadRequest,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		updatedBankAccount, _, err := c.Services.UpdateBankAccountAndlinkStatus(ctx, parsedBankAccountID, request)
		if err != nil{
			dto.Fail(
				ctx,
				http.StatusInternalServerError,
				"ERROR_IN_UPDATING",
				err.Error(),
			)
			return
		}

		dto.Respond(
			ctx,
			http.StatusOK,
			updatedBankAccount,
		)
	}
}

func (c *Controller) UpdateBankAccountBalance() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		var request *dto.UpdateBankAccountBalance
		if err := ctx.ShouldBindJSON(&request); err!= nil{
			dto.Fail(
				ctx,
				http.StatusBadRequest,
				"BINDING_ERROR",
				err.Error(),
			)
			return
		}

		bankAccountID := ctx.Param("bank_account_id")
		parsedBankAccountID, errP := uuid.Parse(bankAccountID)
		if errP != nil{
			dto.Fail(
				ctx,
				http.StatusBadRequest,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		updatedBankAccount, err := c.Services.UpdateBankAccountBalance(ctx, parsedBankAccountID, request)
		if err != nil{
			dto.Fail(
				ctx,
				http.StatusInternalServerError,
				"ERROR_IN_UPDATING",
				err.Error(),
			)
			return
		}

		dto.Respond(
			ctx,
			http.StatusOK,
			updatedBankAccount,
		)
	}
}
