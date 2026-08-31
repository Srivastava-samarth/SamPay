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
	ComplianceService   *services.ComplianceService
	WalletService       *services.WalletService
	MerchantService     *services.MerchantService
	AuthService         *services.AuthService
	UserService         *services.UserService
	MerchantUserService *services.MerchantUserService
	MerchantRepo        repositories.MerchantRepository
}

func NewRegistry(
	db *gorm.DB,
	notificationService *notifications.EmailService,
	merchantSrvc *services.MerchantService,
	complianceSrvc *services.ComplianceService,
	authSrvc *services.AuthService,
	userSrvc *services.UserService,
	merchantUserService *services.MerchantUserService,
	merchantRepo        repositories.MerchantRepository,
) *Registry {
	return &Registry{
		DB:                  db,
		NotificationService: notificationService,
		MerchantService:     merchantSrvc,
		ComplianceService:   complianceSrvc,
		AuthService:         authSrvc,
		UserService:         userSrvc,
		MerchantUserService: merchantUserService,
		MerchantRepo: merchantRepo,
	}
}
