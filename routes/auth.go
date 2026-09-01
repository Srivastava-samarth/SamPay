package routes

import (
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/gin-gonic/gin"
)

type AuthRouter struct{
	AuthController *controllers.AuthController
}

func NewAuthRouter(
	authController *controllers.AuthController,
) *AuthRouter {
	return &AuthRouter{
		AuthController: authController,
	}
}

func(ar *AuthRouter) AuthRoutes(
	router *gin.RouterGroup,
) {
	router.POST(
		"/login",
		ar.AuthController.Login(),
	)
	router.POST(
		"/forgot-password",
		ar.AuthController.ForgotPassword(),
	)
	router.POST(
		"/reset-password",
		ar.AuthController.ResetPassword(),
	)
	router.POST(
		"/refresh-token",
		ar.AuthController.RefreshToken(),
	)
}