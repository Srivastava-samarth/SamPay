package routes

import (
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

func (pr *Router) PayoutRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	payout := router.Group("")
	payout.Use(authMiddleware)
	payout.POST(
		"/:merchant_id/payout/wallet-to-bank",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		pr.Controller.WalletToBankAccount(),
	)
	payout.POST(
		"/:merchant_id/payout/bank-to-bank",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		pr.Controller.BankToBankAccount(),
	)
	payout.GET(
		"/:merchant_id/payouts",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		middlewares.RequireMerchantAccess("super_admin"),
		pr.Controller.GetPayoutsByMerchantID(),
	)
	payout.GET(
		"/:merchant_id/payout/:payout_id",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		middlewares.RequireMerchantAccess("super_admin"),
		pr.Controller.GetPayoutByID(),
	)
}
