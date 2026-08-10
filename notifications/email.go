package notifications

import (
	"fmt"
	"net/smtp"

	"github.com/Srivastava-samarth/sampay/config"
	"github.com/Srivastava-samarth/sampay/dto"
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

