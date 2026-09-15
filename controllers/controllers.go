package controllers

import (
	"github.com/Srivastava-samarth/sampay/services"
	"go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

type Controller struct {
	DB             *gorm.DB
	TemporalClient client.Client

	Services *services.Services
}

func NewController(
	db *gorm.DB,
	temporalClient client.Client,
	service *services.Services,
) *Controller {
	return &Controller{
		DB:             db,
		TemporalClient: temporalClient,
		Services:       service,
	}
}
