package activities

import (
	"gorm.io/gorm"

	"github.com/Srivastava-samarth/sampay/notifications"
)

type Registry struct {
	DB                  *gorm.DB
	NotificationService *notifications.EmailService
}

func NewRegistry(
	db *gorm.DB,
	notificationService *notifications.EmailService,
) *Registry {
	return &Registry{
		DB:                  db,
		NotificationService: notificationService,
	}
}