package main

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/Srivastava-samarth/sampay/database"
	"github.com/Srivastava-samarth/sampay/notifications"
	"github.com/Srivastava-samarth/sampay/routes"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/gin-gonic/gin"
)

func main(){
	logger := utils.NewLogger()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	 notificationService, err := notifications.NewEmailService(cfg.SMTP)

    // router := routes.SetupRoutes(db, emailService)

	router := gin.Default()
	api := router.Group("/api/v1")

	// Register routes
	routes.MerchantRoutes(
		api,
		db,
		notificationService,
	)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
	logger.Println("Starting server on :8080")
	router.Run(":8080")
}