package activities

import (
	"gorm.io/gorm"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/Srivastava-samarth/sampay/nats"
	"github.com/Srivastava-samarth/sampay/notifications"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/services"
)

type Registry struct {
	DB                  *gorm.DB
	InternalConfig      *config.InternalConfig
	NotificationService *notifications.EmailService
	Services            *services.Services
	Repo                *repositories.Repository
	nats                *nats.NatsClient
}

func NewRegistry(
	db *gorm.DB,
	config *config.InternalConfig,
	notificationService *notifications.EmailService,
	services *services.Services,
	repo *repositories.Repository,
	nats *nats.NatsClient,
) *Registry {
	return &Registry{
		DB:                  db,
		InternalConfig:      config,
		NotificationService: notificationService,
		Services:            services,
		Repo:                repo,
		nats:                nats,
	}
}
