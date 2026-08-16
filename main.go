package main

import (
	"net/http"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/Srivastava-samarth/sampay/database"
	"github.com/Srivastava-samarth/sampay/notifications"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/routes"
	"github.com/Srivastava-samarth/sampay/services"
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

	merchantRepo := repositories.NewMerchantRepository(db)
	walletRepo := repositories.NewWalletRepository(db)
	bankRepo := repositories.NewBankRepository(db)
	userRepo := repositories.NewUserRepository(db)
	userSessionRepo := repositories.NewUserSessionRepository(db)
	passwordResetRepo := repositories.NewPasswordResetTokenRepository(db)
	merchantUserRepo := repositories.NewMerchantUserRepository(db)
	linkedBankAccountRepo := repositories.NewLinkedBankRepository(db)

	notificationService, err :=
		notifications.NewEmailService(cfg.SMTP)

	if err != nil {
		logger.Fatalf("Failed to initialize email service: %v", err)
	}

	temporalClient, err := temporal.NewClient(&cfg.Temporal)
	if err != nil {
		logger.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer temporalClient.Close()

	walletService := services.NewWalletService(
		walletRepo,
	)

	bankService := services.NewBankService(
		bankRepo,
		merchantRepo,
		linkedBankAccountRepo,
	)

	userService := services.NewUserService(
		userRepo,
	)

	merchantUserService := services.NewMerchantUserService(
		merchantUserRepo,
	)

	linkedBankAccountService :=
		services.NewLinkedBankAccountService(
			linkedBankAccountRepo,
		)

	complianceService := services.NewComplianceService(
		userRepo,
	)

	merchantService := services.NewMerchantService(
		db,
		merchantRepo,
		bankService,
		complianceService,
		merchantUserService,
		linkedBankAccountService,
		userService,
		walletService,
		notificationService,
	)

	jwtService := utils.NewJwt(&cfg.JWT)

	authService := services.NewAuthService(
		userRepo,
		userSessionRepo,
		merchantUserRepo,
		jwtService,
		notificationService,
		passwordResetRepo,
	)

	// --------------------------------------------------
	// Controllers
	// --------------------------------------------------

	merchantController := controllers.NewMerchantController(
		merchantService,
		temporalClient,
	)

	merchantRouter := routes.NewMerchantRouter(
		merchantController,
	)

	authController := controllers.NewAuthController(
		authService,
		temporalClient,
	)

	authRouter := routes.NewAuthRouter(
		authController,
	)

	activityRegistry := activities.NewRegistry(
		db,
		notificationService,
		merchantService,
		complianceService,
		authService,
	)

	workers := temporal.StartWorkers(
		temporalClient,
		activityRegistry,
	)

	for _, w := range workers {
		go func(w worker.Worker) {
			if err := w.Run(worker.InterruptCh()); err != nil {
				logger.Fatalf("Temporal worker failed: %v", err)
			}
		}(w)
	}

	router := gin.Default()

	api := router.Group("/api/v1")

	merchantRouter.MerchantRoutes(api)
	authRouter.AuthRoutes(api)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	logger.Println("Starting server on :8080")

	if err := router.Run(":8080"); err != nil {
		logger.Fatalf("Server failed: %v", err)
	}
}
