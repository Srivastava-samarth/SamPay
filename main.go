package main

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/Srivastava-samarth/sampay/database"
	"github.com/Srivastava-samarth/sampay/notifications"
	"github.com/Srivastava-samarth/sampay/routes"
	"github.com/Srivastava-samarth/sampay/temporal"
	"github.com/Srivastava-samarth/sampay/temporal/activities"
	"github.com/Srivastava-samarth/sampay/utils"
	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/worker"
)

func main() {
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

	// 4. Create Temporal client
	temporalClient, err := temporal.NewClient(&cfg.Temporal)
	if err != nil {
		logger.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer temporalClient.Close()

	// 5. Create activity registry
	activityRegistry := activities.NewRegistry(
		db,
		notificationService,
	)

	// 6. Start Temporal worker
	temporalWorker := temporal.StartWorker(
		temporalClient,
		activityRegistry,
	)

	go func() {
		if err := temporalWorker.Run(worker.InterruptCh()); err != nil {
			logger.Fatalf("Temporal worker failed: %v", err)
		}
	}()

	// router := routes.SetupRoutes(db, emailService)

	router := gin.Default()
	api := router.Group("/api/v1")

	// Register routes
	routes.MerchantRoutes(
		api,
		db,
		temporalClient,
	)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
	logger.Println("Starting server on :8080")
	router.Run(":8080")
}
