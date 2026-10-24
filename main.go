package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/Srivastava-samarth/sampay/controllers"
	"github.com/Srivastava-samarth/sampay/database"
	"github.com/Srivastava-samarth/sampay/libs"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/Srivastava-samarth/sampay/nats"
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

	sampayPrivateKey, errD := base64.StdEncoding.DecodeString(cfg.Internal.SampayPrivateKey)
	if errD != nil{
		log.Fatalf("decode SamPay private key: %v", err)
	}
	bankPublicKey,errD := base64.StdEncoding.DecodeString(cfg.Internal.BankPublicKey)
	if errD != nil{
		log.Fatalf("decode bank public key: %v", err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}

	repo := repositories.NewRepository(db)

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

	jwtService := middlewares.NewJwt(&cfg.JWT)

	libs := libs.NewLibs(&cfg.Internal)

	services := services.NewServices(db, repo, libs, jwtService, notificationService)

	controllers := controllers.NewController(db, temporalClient, services)
	natsClient, err := nats.NewClient(
		&cfg.Nats,
		services,
		repo,
		temporalClient,
		sampayPrivateKey,
		bankPublicKey,
	)

	if err != nil {
		logger.Fatalf("Failed to create NATS client: %v", err)
	}
	defer natsClient.Close()

	activityRegistry := activities.NewRegistry(
		db,
		&cfg.Internal,
		notificationService,
		services,
		repo,
		natsClient,
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

	ctx, stop := signal.NotifyContext(
    context.Background(),
    os.Interrupt,
    syscall.SIGTERM,
)
defer stop()

	bankingConsumer, err :=
		natsClient.CreateBankingResponsesConsumer(ctx)
	if err != nil {
		logger.Fatalf("Failed to create banking consumer: %v", err)
	}

	consumeContext, err :=
		natsClient.ConsumeBankingCommands(ctx, bankingConsumer)
	if err != nil {
		logger.Fatalf("Failed to start banking consumer: %v", err)
	}

	defer consumeContext.Stop()

	logger.Println("NATS banking consumer started")

	router := gin.Default()

	api := router.Group("/api/v1")

	mainRouter := routes.NewRouter(controllers)

	mainRouter.MerchantRoutes(api, jwtService.Authenticate())
	mainRouter.AuthRoutes(api)
	mainRouter.BankAccountRoutes(api, jwtService.Authenticate())
	mainRouter.UserRoutes(api, jwtService.Authenticate())
	mainRouter.WalletRoutes(api, jwtService.Authenticate())
	mainRouter.VaultRoutes(api, jwtService.Authenticate())
	mainRouter.PaymentRoutes(api, jwtService.Authenticate())
	mainRouter.PayoutRoutes(api, jwtService.Authenticate())
	mainRouter.ReconRoutes(api, jwtService.Authenticate())
	mainRouter.RefundRoutes(api, jwtService.Authenticate())

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
