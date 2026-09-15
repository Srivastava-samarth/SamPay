package routes

import (
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

func (mr *Router) MerchantRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	merchant := router.Group("/merchants")
	merchant.Use(authMiddleware)

	merchant.POST(
		"",
		middlewares.RequireRole("super_admin"),
		mr.Controller.CreateMerchant(),
	)
	merchant.GET(
		"/:merchant_id",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		mr.Controller.GetMerchantByID(),
	)
	merchant.GET(
		"",
		middlewares.RequireRole("super_admin"),
		mr.Controller.GetMerchants(),
	)
	merchant.PATCH(
		"/:merchant_id",
		middlewares.RequireRole("owner", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		mr.Controller.UpdateMerchantInfo(),
	)
	merchant.PATCH(
		"/:merchant_id/update-kyc",
		middlewares.RequireRole("super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		mr.Controller.UpdateMerchantKyc(),
	)
	merchant.PUT(
		"/:merchant_id/re-attempt-kyc",
		middlewares.RequireRole("super_admin"),
		mr.Controller.ReattemptOnboardingKyc(),
	)
	merchant.PATCH(
		"/:merchant_id/status",
		middlewares.RequireRole("super_admin"),
		mr.Controller.UpdateMerchantStatus(),
	)
}
