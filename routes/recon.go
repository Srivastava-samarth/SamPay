package routes

import (
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

func (rr *Router) ReconRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	recon := router.Group("")
	recon.Use(authMiddleware)
	recon.POST(
		"/recon",
		middlewares.RequireRole("super_admin"),
		rr.Controller.TriggerRecon(),
	)
}
