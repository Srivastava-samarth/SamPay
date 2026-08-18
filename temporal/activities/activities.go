package activities

import (
	"gorm.io/gorm"

	"github.com/Srivastava-samarth/sampay/notifications"
	"github.com/Srivastava-samarth/sampay/services"
)

type Registry struct {
	DB                  *gorm.DB
	NotificationService *notifications.EmailService
	ComplianceService   *services.ComplianceService
	WalletService       *services.WalletService
	MerchantService     *services.MerchantService
	AuthService         *services.AuthService
}

func NewRegistry(
	db *gorm.DB,
	notificationService *notifications.EmailService,
	merchantSrvc *services.MerchantService,
	complianceSrvc *services.ComplianceService,
	authSrvc *services.AuthService,
) *Registry {
	return &Registry{
		DB:                  db,
		NotificationService: notificationService,
		MerchantService:     merchantSrvc,
		ComplianceService:   complianceSrvc,
		AuthService: authSrvc,
	}
}
