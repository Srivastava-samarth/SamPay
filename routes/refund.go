package routes

import (
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

func (rr *Router) RefundRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	refund := router.Group("")
	refund.Use(authMiddleware)
	refund.POST(
		"/:merchant_id/refund",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		rr.Controller.CreateRefund(),
	)
	refund.GET(
		"/:merchant_id/refund/:refund_reference",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		middlewares.RequireMerchantAccess("super_admin"),
		rr.Controller.GetRefundByReference(),
	)
	refund.GET(
		"/:merchant_id/refund",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		middlewares.RequireMerchantAccess("super_admin"),
		rr.Controller.GetRefundByMerchantId(),
	)
	refund.GET(
		"/:merchant_id/refunds/:refund_id",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		middlewares.RequireMerchantAccess("super_admin"),
		rr.Controller.GetRefundById(),
	)
}
