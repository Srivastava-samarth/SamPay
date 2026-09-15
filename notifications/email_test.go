package notifications

import (
	"testing"

	"github.com/Srivastava-samarth/sampay/config"
)

func TestNewEmailService(t *testing.T) {
	cfg := config.SMTPConfig{}

	service, err := NewEmailService(cfg)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if service == nil {
		t.Fatal("expected email service, got nil")
	}
}
