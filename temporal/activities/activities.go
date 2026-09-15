package activities

import (
	"gorm.io/gorm"

	"github.com/Srivastava-samarth/sampay/notifications"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"github.com/Srivastava-samarth/sampay/services"
)

type Registry struct {
	DB                  *gorm.DB
	NotificationService *notifications.EmailService
	Services            *services.Services
	Repo                *repositories.Repository
}

func NewRegistry(
	db *gorm.DB,
	notificationService *notifications.EmailService,
	services *services.Services,
	repo *repositories.Repository,
) *Registry {
	return &Registry{
		DB:                  db,
		NotificationService: notificationService,
		Services:            services,
		Repo:                repo,
	}
}
