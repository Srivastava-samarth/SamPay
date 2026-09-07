package routes

import (
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

type PayoutRouter struct {
	payoutCtlr *controllers.PayoutController
}

func NewPayoutRouter(
	payoutCtlr *controllers.PayoutController,
) *PayoutRouter {
	return &PayoutRouter{
		payoutCtlr: payoutCtlr,
	}
}

func (pr *PayoutRouter) PayoutRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	payout := router.Group("")
	payout.Use(authMiddleware)
	payout.POST(
		"/:merchant_id/payout/wallet-to-bank",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		pr.payoutCtlr.WalletToBankAccount(),
	)
	payout.POST(
		"/:merchant_id/payout/bank-to-bank",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		pr.payoutCtlr.BankToBankAccount(),
	)
}