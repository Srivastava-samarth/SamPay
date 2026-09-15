package routes

import (
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

func (pr *Router) PaymentRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	payment := router.Group("")
	payment.Use(authMiddleware)
	payment.POST(
		"/:merchant_id/payment",
		middlewares.RequireRole("owner", "finance"),
		middlewares.RequireMerchantAccess(""),
		pr.Controller.CreatePayment(),
	)
	payment.POST(
		"/settlement",
		middlewares.RequireRole("super_admin"),
		pr.Controller.TriggerSettlement(),
	)
	payment.GET(
		"/:merchant_id/payment",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		middlewares.RequireMerchantAccess("super_admin"),
		pr.Controller.GetPaymentsByMerchantID(),
	)
	payment.GET(
		"/:merchant_id/payment/:payment_id",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		middlewares.RequireMerchantAccess("super_admin"),
		pr.Controller.GetPaymentByID(),
	)
}
