package services

import (
	"github.com/Srivastava-samarth/sampay/libs"
	"github.com/Srivastava-samarth/sampay/middlewares"
	"github.com/Srivastava-samarth/sampay/notifications"
	repositories "github.com/Srivastava-samarth/sampay/respositories"
	"gorm.io/gorm"
)

type Services struct {
	DB *gorm.DB

	Repo                *repositories.Repository
	Libs                *libs.Libs
	JwtService          *middlewares.Jwt
	NotificationService *notifications.EmailService
}

func NewServices(
	db *gorm.DB,
	repo *repositories.Repository,
	libs *libs.Libs,
	jwtService *middlewares.Jwt,
	notificationService *notifications.EmailService,
) *Services {
	return &Services{
		DB:                  db,
		JwtService:          jwtService,
		Libs:                libs,
		NotificationService: notificationService,
		Repo:                repo,
	}
}
