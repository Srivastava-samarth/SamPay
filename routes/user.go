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
	user := router.Group("/user")
	user.Use(authMiddleware)
	user.POST(
		"",
		middlewares.RequireRole("super_admin","owner"),
		ur.UserController.UserOnboarding(),
	)
}