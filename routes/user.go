package routes

import (
	"github.com/Srivastava-samarth/sampay/controllers"
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
) {
	router.POST(
		"/user",
		ur.UserController.UserOnboarding(),
	)
}