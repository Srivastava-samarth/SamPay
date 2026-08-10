package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/Srivastava-samarth/sampay/dto"
)

type MerchantOnboardingWorkflowStarter interface {
	StartMerchantOnboarding(
		ctx context.Context,
		request dto.CreateMerchantOnboardingRequest,
		merchantID uuid.UUID,
	) error
}