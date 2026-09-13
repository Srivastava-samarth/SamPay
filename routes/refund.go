package routes

import (
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

type RefundRouter struct {
	refundCtrl *controllers.RefundController
}

func NewRefundRouter(
	refundCtrl *controllers.RefundController,
) *RefundRouter {
	return &RefundRouter{
		refundCtrl: refundCtrl,
	}
}

func (rr *RefundRouter) RefundRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	refund := router.Group("")
	refund.Use(authMiddleware)
	refund.POST(
		"/:merchant_id/refund",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		rr.refundCtrl.CreateRefund(),
	)
	refund.GET(
		"/:merchant_id/refund/:refund_reference",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		middlewares.RequireMerchantAccess("super_admin"),
		rr.refundCtrl.GetRefundByReference(),
	)
	refund.GET(
		"/:merchant_id/refund",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		middlewares.RequireMerchantAccess("super_admin"),
		rr.refundCtrl.GetRefundByMerchantId(),
	)
	refund.GET(
		"/:merchant_id/refunds/:refund_id",
		middlewares.RequireRole("super_admin", "owner", "finance"),
		middlewares.RequireMerchantAccess("super_admin"),
		rr.refundCtrl.GetRefundById(),
	)
}
