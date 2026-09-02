package routes

import (
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

type MerchantRouter struct {
	MerchantController *controllers.MerchantController
}

func NewMerchantRouter(
	merchantController *controllers.MerchantController,
) *MerchantRouter {
	return &MerchantRouter{
		MerchantController: merchantController,
	}
}

func (mr *MerchantRouter) MerchantRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	merchant := router.Group("/merchants")
	merchant.Use(authMiddleware)

	merchant.POST(
		"",
		middlewares.RequireRole("super_admin"),
		mr.MerchantController.CreateMerchant(),
	)
}