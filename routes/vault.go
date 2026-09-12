package routes

import (
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/gin-gonic/gin"
)

type VaultRouter struct {
	vaultCtlr *controllers.VaultController
}

func NewVaultRouter(
	vaultCtlr *controllers.VaultController,
) *VaultRouter{
	return &VaultRouter{
		vaultCtlr: vaultCtlr,
	}
}

func (vr *VaultRouter) VaultRoutes(
	router *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
) {
	vault := router.Group("/vault")
	vault.Use(authMiddleware)
	vault.POST(
		"",
		middlewares.RequireRole("super_admin"),
		vr.vaultCtlr.CreateVault(),
	)
	vault.GET(
		"",
		middlewares.RequireRole("super_admin"),
		vr.vaultCtlr.GetVaults(),
	)
	vault.GET(
		"/:vault_id",
		middlewares.RequireRole("super_admin"),
		vr.vaultCtlr.GetVault(),
	)
	vault.PATCH(
		"/balance",
		middlewares.RequireRole("super_admin"),
		vr.vaultCtlr.UpdateVaultBalance(),
	)
	vault.PATCH(
		"/:vault_id/status",
		middlewares.RequireRole("super_admin"),
		vr.vaultCtlr.UpdateVaultStatus(),
	)
}