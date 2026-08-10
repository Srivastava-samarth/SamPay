package routes

import (
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

func MerchantRoutes(
	router *gin.RouterGroup,
	db *gorm.DB,
	temporalClient client.Client,
) {
	router.POST(
		"/merchants",
		controllers.CreateMerchant(db, temporalClient),
	)
}