package dto

type AuthRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type ForgotPasswordRequest struct{
	Email string `json:"email"`
}

type ResetPasswordRequest struct{
	Email string `json:"email" validate:"required"`
	ResetToken string `json:"reset_token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}
