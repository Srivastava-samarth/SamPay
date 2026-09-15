package routes

import (
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

func (vr *Router) VaultRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	vault := router.Group("/vault")
	vault.Use(authMiddleware)
	vault.POST(
		"",
		middlewares.RequireRole("super_admin"),
		vr.Controller.CreateVault(),
	)
	vault.GET(
		"",
		middlewares.RequireRole("super_admin"),
		vr.Controller.GetVaults(),
	)
	vault.GET(
		"/:vault_id",
		middlewares.RequireRole("super_admin"),
		vr.Controller.GetVault(),
	)
	vault.PATCH(
		"/balance",
		middlewares.RequireRole("super_admin"),
		vr.Controller.UpdateVaultBalance(),
	)
	vault.PATCH(
		"/:vault_id/status",
		middlewares.RequireRole("super_admin"),
		vr.Controller.UpdateVaultStatus(),
	)
}
