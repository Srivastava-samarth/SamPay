package controllers

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/constants"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (wc *Controller) GetWallet() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantID := c.Param("merchant_id")
		if merchantID == "" {
			dto.Fail(
				c,
				http.StatusNotFound,
				"MERCHANT_ID_NOT_FOUND",
				"Merchant ID not passed in params",
			)
			return
		}

		parsedMerchantId, errP := uuid.Parse(merchantID)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		wallet, errW := wc.WalletService.GetWalletByMerchantID(parsedMerchantId)
		if errW != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"WALLET_NOT_FOUND",
				errW.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			wallet,
		)
	}
}

func (wc *Controller) GetWalletTransactions() gin.HandlerFunc {
	return func(c *gin.Context) {

		merchantID := c.Param("merchant_id")
		accountType := constants.LedgerAccountTypeWallet

		if merchantID == "" || accountType == "" {
			dto.Fail(
				c,
				http.StatusNotFound,
				"MERCHANT_ID_OR_ACCOUNT_TYPE_NOT_FOUND",
				"Merchant ID or Account type not passed in request",
			)
			return
		}

		parsedMerchantID, err := uuid.Parse(merchantID)
		if err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"PARSING_ERROR",
				err.Error(),
			)
			return
		}

		wallet, err := wc.WalletService.GetWalletByMerchantID(parsedMerchantID)
		if err != nil {
			dto.Fail(
				c,
				http.StatusNotFound,
				"MERCHANT_ID_NOT_FOUND",
				"merchant id not found",
			)
			return
		}

		cursorParam := c.Query("cursor")
		direction := c.DefaultQuery("direction", "next")

		var cursor *uuid.UUID

		if cursorParam != "" {
			parsedCursor, err := uuid.Parse(cursorParam)
			if err != nil {
				dto.Fail(
					c,
					http.StatusBadRequest,
					"INVALID_CURSOR",
					"Invalid cursor",
				)
				return
			}

			cursor = &parsedCursor
		}

		transactions, pagination, err := wc.LedgerService.GetTransactions(
			&accountType,
			wallet.ID,
			cursor,
			direction,
		)

		if err != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"TRANSACTIONS_NOT_FOUND_ISSUE",
				err.Error(),
			)
			return
		}

		var walletTransactions []dto.WalletTransactionResponse

		for _, t := range transactions {
			walletTransactions = append(
				walletTransactions,
				dto.WalletTransactionResponse{
					ID:               t.LedgerTransactionID,
					TransactionRef:   t.TransactionRef,
					Type:             t.Type,
					EntryType:        t.EntryType,
					ReferenceID:      t.ReferenceID,
					Amount:           t.Amount,
					Currency:         t.Currency,
					Status:           t.Status,
					SettlementStatus: t.SettlementStatus,
					CreatedAt:        t.CreatedAt,
				},
			)
		}

		dto.Respond(
			c,
			http.StatusOK,
			gin.H{
				"data":       walletTransactions,
				"pagination": pagination,
			},
		)
	}
}

func (wc *Controller) GetWalletTransaction() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantID := c.Param("merchant_id")
		transactionId := c.Param("transaction_id")
		if merchantID == "" {
			dto.Fail(
				c,
				http.StatusNotFound,
				"MERCHANT_ID_NOT_FOUND",
				"Merchant ID not passed in params",
			)
			return
		}

		if transactionId == "" {
			dto.Fail(
				c,
				http.StatusNotFound,
				"TRANSACTION_ID_NOT_FOUND",
				"Transaction ID not passed in params",
			)
			return
		}

		parsedMerchantId, errP := uuid.Parse(merchantID)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		parsedTransactionId, errP := uuid.Parse(transactionId)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		wallet, errW := wc.WalletService.GetWalletByMerchantID(parsedMerchantId)
		if errW != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"WALLET_NOT_FOUND",
				errW.Error(),
			)
			return
		}

		transaction, errT := wc.LedgerService.GetTransaction(wallet.ID, parsedTransactionId)
		if errT != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"TRANSACTIONS_NOT_FOUND_ISSUE",
				errT.Error(),
			)
			return
		}
		dto.Respond(
			c,
			http.StatusOK,
			transaction,
		)
	}
}

func (wc *Controller) TopUpWallet() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request *dto.TopupWalletBalanceRequest
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

		err := c.ShouldBindJSON(&request)
		if err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"BINDING_ERROR",
				err.Error(),
			)
			return
		}

		wallet, errW := wc.WalletService.GetWalletByMerchantID(parsedMerchantID)
		if errW != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_FINDING_WALLET",
				errW.Error(),
			)
		}

		updatedWallet, errUW := wc.WalletService.TopUpWalletFromPrimaryBank(wc.DB, wallet, request.Amount)
		if errUW != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_TOPING_UP_WALLET",
				errUW.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			updatedWallet,
		)
	}
}
