package routes

import (
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

func (br *Router) BankAccountRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	bankAccount := router.Group("/merchant")
	bankAccount.Use(authMiddleware)

	bankAccount.GET(
		"/:merchant_id/bank_account/:bank_account_id",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		br.Controller.GetBankAccount(),
	)

	bankAccount.GET(
		"/:merchant_id/bank_accounts",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		br.Controller.GetBankAccounts(),
	)
	bankAccount.POST(
		"/:merchant_id/bank_account",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		br.Controller.CreateBankAccount(),
	)
	bankAccount.PATCH(
		"/:merchant_id/bank_account/:bank_account_id",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		br.Controller.UpdateBankAccountStatus(),
	)
	bankAccount.PATCH(
		"/:merchant_id/bank_account/:bank_account_id/balance",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		br.Controller.UpdateBankAccountBalance(),
	)
}