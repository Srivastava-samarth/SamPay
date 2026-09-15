package routes

import (
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

func (ur *Router) UserRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	user := router.Group("/:merchant_id/user")
	user.Use(authMiddleware)
	user.POST(
		"",
		middlewares.RequireRole("super_admin", "owner"),
		middlewares.RequireMerchantAccess("super_admin"),
		ur.Controller.UserOnboarding(),
	)
	user.GET(
		"/:user_id",
		middlewares.RequireRole("super_admin", "owner"),
		middlewares.RequireMerchantAccess("super_admin"),
		ur.Controller.GetUserByID(),
	)
	user.GET(
		"/merchant",
		middlewares.RequireRole("super_admin", "owner"),
		middlewares.RequireMerchantAccess("super_admin"),
		ur.Controller.GetUsersByMerchant(),
	)
	user.GET(
		"",
		middlewares.RequireMerchantAccess("super_admin"),
		ur.Controller.GetUsers(),
	)
	user.PATCH(
		"/:user_id/status",
		middlewares.RequireRole("super_admin", "owner"),
		middlewares.RequireMerchantAccess("super_admin"),
		ur.Controller.UpdateUserStatus(),
	)
}
