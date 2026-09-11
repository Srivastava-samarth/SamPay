package routes

import (
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

type BankRouter struct {
	bankCtlr *controllers.BankController
}

func NewBankRouter(
	bankAccountController *controllers.BankController,
) *BankRouter {
	return &BankRouter{
		bankCtlr: bankAccountController,
	}
}

func (br *BankRouter) BankAccountRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	bankAccount := router.Group("/merchant")
	bankAccount.Use(authMiddleware)

	bankAccount.POST(
		"/:merchant_id/bank_account",
		middlewares.RequireRole("owner", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		br.bankCtlr.CreateBankAccount(),
	)
	bankAccount.GET(
		"/:merchant_id/bank_account/:bank_account_id",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		br.bankCtlr.GetBankAccount(),
	)

	bankAccount.GET(
		"/:merchant_id/bank_accounts",
		middlewares.RequireRole("owner", "finance", "super_admin"),
		middlewares.RequireMerchantAccess("super_admin"),
		br.bankCtlr.GetBankAccounts(),
	)
	bankAccount.PATCH(
		"/:merchant_id/bank_account",
		middlewares.RequireRole("super_admin"),
		br.bankCtlr.UpdateBankAccount(),
	)
}
