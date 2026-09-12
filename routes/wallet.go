package routes

import (
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

type WalletRouter struct {
	walletCntlr *controllers.WalletController
}

func NewWalletRouter(
	walletController *controllers.WalletController,
) *WalletRouter {
	return &WalletRouter{
		walletCntlr: walletController,
	}
}

func (wr *WalletRouter) WalletRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	wallet := router.Group("/merchant")
	wallet.Use(authMiddleware)

	wallet.GET(
		"/:merchant_id/wallet",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		wr.walletCntlr.GetWallet(),
	)

	wallet.GET(
		"/:merchant_id/wallet/transactions",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		wr.walletCntlr.GetWalletTransactions(),
	)
	wallet.GET(
		"/:merchant_id/wallet/transaction/:transaction_id",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		wr.walletCntlr.GetWalletTransaction(),
	)
}
