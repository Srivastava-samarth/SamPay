package routes

import (
	"github.com/gin-gonic/gin"
)

func (ar *Router) AuthRoutes(
	router *gin.RouterGroup,
) {
	router.POST(
		"/login",
		ar.Controller.Login(),
	)
	router.POST(
		"/forgot-password",
		ar.Controller.ForgotPassword(),
	)
	router.POST(
		"/reset-password",
		ar.Controller.ResetPassword(),
	)
	router.POST(
		"/refresh-token",
		ar.Controller.RefreshToken(),
	)
}
