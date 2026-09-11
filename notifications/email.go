package notifications

import (
	"fmt"
	"net/smtp"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/Srivastava-samarth/sampay/dto"
	"github.com/google/uuid"
)

type EmailService struct {
	Config config.SMTPConfig
}

func NewEmailService(cfg config.SMTPConfig) (*EmailService, error) {
	return &EmailService{
		Config: cfg,
	}, nil
}

func (e *EmailService) Send(
	to string,
	subject string,
	body string,
) error {

	address := fmt.Sprintf(
		"%s:%s",
		e.Config.Host,
		e.Config.Port,
	)

	message := []byte(
		"From: " + e.Config.From + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			body,
	)

	return smtp.SendMail(
		address,
		nil,
		e.Config.From,
		[]string{to},
		message,
	)
}

func (s *EmailService) SendMerchantWelcomeEmail(
	user *dto.CreateUserResponse,
	temporaryPassword string,
) error {

	subject := "Welcome to SamPay"

	body := fmt.Sprintf(`
Hello %s,

Your SamPay merchant account has been successfully created.

Login Email: %s
Temporary Password: %s

You must change your password after at your first login.

Regards,
SamPay Team
`, user.FirstName, user.Email, temporaryPassword)

	return s.Send(
		user.Email,
		subject,
		body,
	)
}

func (s *EmailService) SendResetPasswordEmail(
	name string,
	email string,
	token string,
) error {
	subject := "SamPay: Forgot Password Reset"

	body := fmt.Sprintf(`
Hello %s,

You have initiated a forgot password request through our platform.

Login Email: %s
reset password access token: %s

The password reset will take place through another API which will use this access token provided to validate the request.
As SamPay is currently a API based service, you may please access our rest password endpoint.

Regards,
SamPay Team
`, name, email, token)

	return s.Send(
		email,
		subject,
		body,
	)
}

func (s *EmailService) SendUserOnboardingEmail(
	name string,
	email string,
	merchantName string,
	role string,
	password string,
) error {
	subject := "Welcome to SamPay"

	body := fmt.Sprintf(`
Hello %s,

You have been onboarded by SamPay merchant %s with role %s.
In order to access the functionalities use the login credentials.

Login Email: %s
Temporary Password: %s

You must change your password after at your first login.

Regards,
SamPay Team
`, name, merchantName, role, email, password)

	return s.Send(
		email,
		subject,
		body,
	)
}

func (s *EmailService) SendReconEmail(
	body string,
) error {

	email := s.Config.From
	subject := "SamPay Daily Reconciliation Report"

	return s.Send(
		email,
		subject,
		body,
	)
}


func (s *EmailService) SendKYCReattemptEmail(
	email string,
	merchantID uuid.UUID,
	merchantType string,
) error {
	subject := "SamPay KYC Verification - Action Required"

	body := fmt.Sprintf(`
Hello,

We were unable to complete the compliance verification for your SamPay merchant account.

You can submit your KYC information again to continue the onboarding process.

Merchant ID: %s
Merchant Type: %s

Please use the Merchant ID above when submitting your KYC information again.

If you believe this was done in error or need assistance, please contact our support team.

Regards,
SamPay Team
`, merchantID.String(), merchantType)

	return s.Send(
		email,
		subject,
		body,
	)
}

