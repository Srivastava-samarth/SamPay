package controllers

import (
	"github.com/Srivastava-samarth/sampay/dto"
	services "github.com/Srivastava-samarth/sampay/services"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"net/http"
)

var validate = validator.New()

func CreateMerchant(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		var request dto.CreateMerchantOnboardingRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				err.Error(),
			)
			return
		}

		if err := validate.Struct(request); err != nil {
			dto.Fail(
				c,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				err.Error(),
			)
			return
		}

		response, err := services.CreateMerchant(request, db)
		if err != nil {
			dto.Fail(
				c,
				http.StatusInternalServerError,
				"MERCHANT_ONBOARDING_FAILED",
				err.Error(),
			)
			return
		}

		dto.Respond(
			c,
			http.StatusAccepted,
			response,
		)
	}
}
