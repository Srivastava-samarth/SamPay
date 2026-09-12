package routes

import (
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

type UserRouter struct {
	UserController *controllers.UserController
}

func NewUserRouter(
	userController *controllers.UserController,
) *UserRouter {
	return &UserRouter{
		UserController: userController,
	}
}

func (ur *UserRouter) UserRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	user := router.Group("/:merchant_id/user")
	user.Use(authMiddleware)
	user.POST(
		"",
		middlewares.RequireRole("super_admin","owner"),
		middlewares.RequireMerchantAccess("super_admin"),
		ur.UserController.UserOnboarding(),
	)
	user.GET(
		"/:user_id",
		middlewares.RequireRole("super_admin","owner"),
		middlewares.RequireMerchantAccess("super_admin"),
		ur.UserController.GetUserByID(),
	)
	user.GET(
		"/merchant",
		middlewares.RequireRole("super_admin","owner"),
		middlewares.RequireMerchantAccess("super_admin"),
		ur.UserController.GetUsersByMerchant(),
	)
	user.GET(
		"",
		middlewares.RequireMerchantAccess("super_admin"),
		ur.UserController.GetUsers(),
	)
	user.PATCH(
		"/:user_id/status",
		middlewares.RequireRole("super_admin","owner"),
		middlewares.RequireMerchantAccess("super_admin"),
		ur.UserController.UpdateUserStatus(),
	)
}