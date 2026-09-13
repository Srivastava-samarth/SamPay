package controllers

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/Srivastava-samarth/sampay/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BankController struct {
	BankSrvc *services.BankService
}

func NewBankController(
	bankSrvc *services.BankService,
) *BankController {
	return &BankController{
		BankSrvc: bankSrvc,
	}
}

func (bc *BankController) CreateBankAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request *dto.CreateBankAccountRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

		bankAccount, linkedBankAccount, err := bc.BankSrvc.CreateBankAccountAndLink(request)
		if err != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"BANK_CREATION_FAILED",
				err.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			map[string]interface{}{
				"bank_account":        bankAccount,
				"linked_bank_account": linkedBankAccount,
			},
		)
	}
}

func (bc *BankController) UpdateBankAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request *dto.UpdateBankAccountRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

		if request.ID == uuid.Nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"id is empty",
				"ID should be passed",
			)
			return
		}

		updatedBankAccount, updateLinkedBankAccount, err := bc.BankSrvc.UpdateBankAccountAndlink(request)
		if err != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"UPDATION_ISSUE",
				err.Error(),
			)
		}

		dto.Respond(
			c,
			http.StatusOK,
			map[string]interface{}{
				"bank_account":        updatedBankAccount,
				"linked_bank_account": updateLinkedBankAccount,
			},
		)
	}
}

func (bc *BankController) GetBankAccounts() gin.HandlerFunc {
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

		bankAccounts, errBA := bc.BankSrvc.GetBankAccountsByMerchantID(parsedMerchantID)
		if errBA != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"BANK_ACCOUNT_NOT_FOUND",
				errBA.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			bankAccounts,
		)
	}
}

func (bc *BankController) GetBankAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		bankAccountID := c.Param("bank_account_id")
		parsedBankAccountID, errP := uuid.Parse(bankAccountID)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		bankAccount, errBA := bc.BankSrvc.GetBankAccount(parsedBankAccountID)
		if errBA != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"BANK_ACCOUNT_NOT_FOUND",
				errBA.Error(),
			)
		}

		dto.Respond(
			c,
			http.StatusOK,
			bankAccount,
		)
	}
}
