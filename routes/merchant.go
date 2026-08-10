package routes

import (
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/Srivastava-samarth/sampay/notifications"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func MerchantRoutes(
	router *gin.RouterGroup,
	db *gorm.DB,
	notificationService *notifications.EmailService,
) {
	router.POST(
		"/merchants",
		controllers.CreateMerchant(db, notificationService),
	)
}