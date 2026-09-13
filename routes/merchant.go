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
	merchant.GET(
		"/:merchant_id",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		mr.MerchantController.GetMerchantByID(),
	)
	merchant.GET(
		"",
		middlewares.RequireRole("super_admin"),
		mr.MerchantController.GetMerchants(),
	)
	merchant.PATCH(
		"/:merchant_id",
		middlewares.RequireRole("owner", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		mr.MerchantController.UpdateMerchantInfo(),
	)
	merchant.PATCH(
		"/:merchant_id/update-kyc",
		middlewares.RequireRole("super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		mr.MerchantController.UpdateMerchantKyc(),
	)
	merchant.PUT(
		"/:merchant_id/re-attempt-kyc",
		middlewares.RequireRole("super_admin"),
		mr.MerchantController.ReattemptOnboardingKyc(),
	)
	merchant.PATCH(
		"/:merchant_id/status",
		middlewares.RequireRole("super_admin"),
		mr.MerchantController.UpdateMerchantStatus(),
	)
}
