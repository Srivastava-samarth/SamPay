package controllers

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (vc *Controller) CreateVault() gin.HandlerFunc {
	return func(c *gin.Context) {
		var vaultRequest *dto.CreateVaultRequest
		if err := c.ShouldBindJSON(&vaultRequest); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

		vault, errV := vc.VaultService.CreateVault(vaultRequest)
		if errV != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_CREATING_VAULT",
				errV.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			vault,
		)
	}
}

func (vc *Controller) GetVault() gin.HandlerFunc {
	return func(c *gin.Context) {
		vaultId := c.Param("vault_id")
		parsedVaultId, errP := uuid.Parse(vaultId)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		vault, errV := vc.VaultService.GetVaultByID(parsedVaultId)
		if errV != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_GETTING_VAULT",
				errV.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			vault,
		)
	}
}

func (vc *Controller) GetVaults() gin.HandlerFunc {
	return func(c *gin.Context) {
		vaults, errV := vc.VaultService.GetVaults()
		if errV != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_GETTING_VAULTS",
				errV.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			vaults,
		)
	}
}

func (vc *Controller) UpdateVaultBalance() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantID, exist := c.Get("merchant_id")
		if !exist {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"MERCHANT_ID_NOT_FOUND",
				"merchant_id not found in context",
			)
			return
		}
		var request *dto.UpdateVaultBalance
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

		vault, errV := vc.VaultService.UpdateVaultBalance(merchantID.(uuid.UUID), request.Balance, request.Type)
		if errV != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_UPDATING_VAULT",
				errV.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			vault,
		)
	}
}

func (vc *Controller) UpdateVaultStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request *dto.UpdateVaultStatus
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

		vaultId := c.Param("vault_id")
		parsedVaultID, errP := uuid.Parse(vaultId)
		if errP != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"PARSING_ERROR",
				errP.Error(),
			)
			return
		}

		updatedVault, errUV := vc.VaultService.UpdateVaultStatus(request.Status, parsedVaultID)
		if errUV != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"ERROR_UPDATING_VAULT",
				errUV.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusOK,
			updatedVault,
		)
	}
}
