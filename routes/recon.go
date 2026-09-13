package routes

import (
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

type ReconRouter struct {
	reconCtlr *controllers.ReconController
}

func NewReconRouter(
	reconCtlr *controllers.ReconController,
) *ReconRouter {
	return &ReconRouter{
		reconCtlr: reconCtlr,
	}
}

func (rr *ReconRouter) ReconRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	recon := router.Group("")
	recon.Use(authMiddleware)
	recon.POST(
		"/recon",
		middlewares.RequireRole("super_admin"),
		rr.reconCtlr.TriggerRecon(),
	)
}
