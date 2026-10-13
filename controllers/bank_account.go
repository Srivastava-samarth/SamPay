package controllers

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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