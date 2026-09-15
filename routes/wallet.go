package routes

import (
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

func (wr *Router) WalletRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	wallet := router.Group("/merchant")
	wallet.Use(authMiddleware)

	wallet.GET(
		"/:merchant_id/wallet",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		wr.Controller.GetWallet(),
	)

	wallet.GET(
		"/:merchant_id/wallet/transactions",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		wr.Controller.GetWalletTransactions(),
	)
	wallet.GET(
		"/:merchant_id/wallet/transaction/:transaction_id",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		wr.Controller.GetWalletTransaction(),
	)
	wallet.PUT(
		"/:merchant_id/wallet/topup",
		middlewares.RequireRole("super_admin"),
		wr.Controller.TopUpWallet(),
	)
}
