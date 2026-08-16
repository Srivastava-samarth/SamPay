package activities

import (
	"context"

	"github.com/Srivastava-samarth/sampay/dto"
)

func (a *Registry) ForgotPassword(
	ctx context.Context,
	request *dto.ForgotPasswordRequest,
) error {
	return a.AuthService.ForgotPasswod(request)
}
