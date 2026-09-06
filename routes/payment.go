package routes

import (
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

type PaymentRouter struct {
	paymentCtlr *controllers.PaymentController
}

func NewPaymentRouter(
	paymentCtlr *controllers.PaymentController,
) *PaymentRouter {
	return &PaymentRouter{
		paymentCtlr: paymentCtlr,
	}
}

func (pr *PaymentRouter) PaymentRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	payment := router.Group("")
	payment.Use(authMiddleware)
	payment.POST(
		"/:merchant_id/payment",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		pr.paymentCtlr.CreatePayment(),
	)
}