package routes

import (
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/gin-gonic/gin"
)

type MerchantRouter struct{
	MerchantController *controllers.MerchantController
}

func NewMerchantRouter(
	merchantController *controllers.MerchantController,
) *MerchantRouter{
	return &MerchantRouter{
		MerchantController: merchantController,
	}
}

func(mr *MerchantRouter) MerchantRoutes(
	router *gin.RouterGroup,
) {
	router.POST(
		"/merchants",
		mr.MerchantController.CreateMerchant(),
	)
}